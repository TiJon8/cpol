package pgconn

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"strconv"
	"time"

	"github.com/TiJon8/cpol/proto"
	"github.com/TiJon8/cpol/utils"
)

const (
	connStatusUninitialized = iota
	connStatusConnecting
	connStatusClosed
	connStatusIdle
	connStatusBusy
)


type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type LookupFunc func(ctx context.Context, host string) (addrs []string, err error)

type FrontendBuilderFunc func(r io.Reader, w io.Writer) *proto.Frontend


type PgConn struct {
	connect net.Conn
	pid uint32
	secretKey []byte

	frontend *proto.Frontend
	config *ConnectionConfig

	parameterStatuses map[string]string
	status byte

	simpleQuery SimpleQuery // to-do
}

func ParseConfig(configMap map[string]string) (*ConnectionConfig, error) {
	config, err := ParseConfigMap(configMap)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ConnectConfig(ctx context.Context, config *ConnectionConfig) (*PgConn, error) {
	if !config.createdByParsingConfig {
		panic("not created by ParseConfig")
	}
	var allErrors []error
	fmt.Println(config)
	connectConfigs, errs := buildConnectOneConfigs(ctx, config)
	if len(errs) > 0 {
		allErrors = append(allErrors, errs...)
	}
	if len(connectConfigs) == 0 {
		return nil, errors.Join(allErrors...)
	}

	pgConn, errs := makeConnect(ctx, config, connectConfigs)
	if len(errs) > 0 {
		allErrors = append(allErrors, errs...)
		return nil, errors.Join(errs...)
	}

	return pgConn, nil
}

type connectOneConfig struct {
	network          string
	address          string
	originalHostname string      // original hostname before resolving
}

func buildConnectOneConfigs(ctx context.Context, config *ConnectionConfig) ([]*connectOneConfig, []error) {
	var errors []error
	var configs []*connectOneConfig
	ips, err := config.LookupFunc(ctx, config.Host)
	fmt.Println(ips)
	if err != nil {
		errors = append(errors, err)
	}
	for _, ip := range ips {
		h, p, err := net.SplitHostPort(ip)
		fmt.Println(h, p, err)
		if err != nil {
			network, address := NetworkAddress(ip, config.Port)
			fmt.Println(network, address)
			configs = append(configs, &connectOneConfig{
				network: network,
				address: address,
				originalHostname: config.Host,
			})
		} else {
			port, err := strconv.ParseUint(p, 10, 16)
			if err != nil {
				return nil, []error{fmt.Errorf("%v", err)}
			}
			network, address := NetworkAddress(h, uint16(port))
			configs = append(configs, &connectOneConfig{
				network: network,
				address: address,
				originalHostname: config.Host,
			})
		}
	}
	fmt.Println(configs)
	return configs, errors
}


func makeConnect(ctx context.Context, config *ConnectionConfig, connectConfig []*connectOneConfig) (*PgConn, []error) {
	oldCtx := ctx
	var allErrors []error
	for i, c := range connectConfig {
		fmt.Println("connectConfig pgConn loop", len(connectConfig))
		if config.ConnectTimeout != 0 {
			// create new context first time or when previous host was different
			if i == 0 || (connectConfig[i].address != connectConfig[i-1].address) {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(oldCtx, config.ConnectTimeout)
				defer cancel()
			}
		} else {
			ctx = oldCtx
		}

		pgConn, err := connectOne(ctx, config, c)
		if pgConn != nil {
			return pgConn, nil
		}
		allErrors = append(allErrors, err)
	}

	return nil, allErrors
}

func connectOne(ctx context.Context, config *ConnectionConfig, connectOneConfig *connectOneConfig) (*PgConn, error) {
	pgConn := new(PgConn)
	pgConn.config = config

	var err error
	pgConn.connect, err = config.DialFunc(ctx, connectOneConfig.network, connectOneConfig.address)
	if err != nil {
		return nil, err
	}

	maxProtocolVersion, err := parseProtocolVersion(config.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	pgConn.parameterStatuses = make(map[string]string)
	pgConn.frontend = config.FrontendBuilder(pgConn.connect, pgConn.connect)

	startupMsg := proto.StartupMessage{
		ProtocolVersion: maxProtocolVersion,
		Parameters:      make(map[string]string),
	}

	maps.Copy(pgConn.parameterStatuses, config.RuntimeParams)

	startupMsg.Parameters["user"] = config.User
	if config.Database != "" {
		startupMsg.Parameters["database"] = config.Database
	}
	pgConn.frontend.Send(&startupMsg)
	if err := pgConn.flush(); err != nil {
		pgConn.connect.Close()
		return nil, err
	}

	for {
		msg, err := pgConn.receiveMessage()
		if err != nil {
			pgConn.connect.Close()
			return nil, fmt.Errorf("failed to receive message: %w", err)
		}
		switch msg := msg.(type) {
		case *proto.AuthenticationOk:
		case *proto.AuthenticationCleartextPassword:
			err = pgConn.txPasswordMessage(pgConn.config.Password)
			if err != nil {
				pgConn.connect.Close()
				return nil, err
			}
		case *proto.ParameterStatus:
		case *proto.BackendKeyData:
			pgConn.pid = msg.ProcessID
			pgConn.secretKey = msg.SecretKey
		case *proto.ReadyForQuery:
			pgConn.status = connStatusIdle
			fmt.Println("ReadyForQuery status", msg.TxStatus)
			return pgConn, nil
		case *proto.ErrorResponse:
			pgConn.connect.Close()
			return nil, fmt.Errorf("%s", msg.Error)
		default:
			pgConn.connect.Close()
			return nil, fmt.Errorf("unexpected message")
		}
	}
}

func (pgConn *PgConn) receiveMessage() (proto.BackendMessage, error) {
	msg, err := pgConn.peekMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (pgConn *PgConn) peekMessage() (proto.BackendMessage, error) {
	var msg proto.BackendMessage
	var err error
	msg, err = pgConn.frontend.Receive()
	if err != nil {
		// Close on anything other than timeout error - everything else is fatal
		var netErr net.Error
		isNetErr := errors.As(err, &netErr)
		if !(isNetErr && netErr.Timeout()) {
			fmt.Println("Closing connection by async close")
			pgConn.asyncClose()
		}

		return nil, err
	}
	return msg, nil
}


func (pgConn *PgConn) asyncClose() {
	if pgConn.status == connStatusClosed {
		return
	}
	pgConn.status = connStatusClosed
	go func() {
		defer pgConn.connect.Close()

		deadline := time.Now().Add(time.Second * 15)

		// async cancel worked request
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		pgConn.CancelRequest(ctx)

		pgConn.connect.SetDeadline(deadline)

		pgConn.frontend.Send(&proto.Terminate{})
		pgConn.flush()
	}()
}


func (pgConn *PgConn) CancelRequest(ctx context.Context) (error) {
	remoteAddr := pgConn.connect.RemoteAddr()
	serverNetwork := remoteAddr.Network()
	serverAddress := remoteAddr.String()
	connect, err := pgConn.config.DialFunc(ctx, serverNetwork, serverAddress)
	if err != nil {
		return err
	}
	defer connect.Close()
	buf := new(bytes.Buffer)
	bufBody := new(bytes.Buffer)
	binary.Write(bufBody, binary.BigEndian, uint32(80877102))
	binary.Write(bufBody, binary.BigEndian, uint32(pgConn.pid))
	bufBody.Write(pgConn.secretKey)
	binary.Write(buf, binary.BigEndian, uint32(bufBody.Len()+4))
	buf.Write(bufBody.Bytes())

	if _, err := connect.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}


func (pgConn *PgConn) flush() error {
	err := pgConn.frontend.Flush()
	return err
}

func (pgConn *PgConn) txPasswordMessage(password string) error {
	pgConn.frontend.Send(&proto.PasswordMessage{Password: password})
	return pgConn.flush()
}


func (pgConn *PgConn) Close(ctx context.Context) error {
	if pgConn.status == connStatusClosed {
		return nil
	}
	pgConn.status = connStatusClosed

	pgConn.frontend.Send(&proto.Terminate{})
	pgConn.flush()

	return pgConn.connect.Close()
}


func NetworkAddress(host string, port uint16) (network, address string) {
	network = "tcp"
	address = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return network, address
}




func (pgConn *PgConn) Ping(ctx context.Context) error {
	return pgConn.Exec(ctx, "-- ping").Close()
}

func (pgConn *PgConn) prepareContext(ctx context.Context) (*SimpleQuery) {
	pgConn.simpleQuery = SimpleQuery{
		pgConn: pgConn,
		ctx:    ctx,
		resultKeeper: &ResultKeeper{},
		quryProcessChan: make(chan struct{}),
	}
	queryResult := &pgConn.simpleQuery

	if ctx != context.Background() {
		go func ()  {
			select {
			case <-ctx.Done():
				queryResult.closed = true
				queryResult.err = fmt.Errorf("context done")
				fmt.Println("Контекст отменен, шлем CancelRequest...")
				select {
				case <-queryResult.quryProcessChan:
					return
				default:
					pgConn.asyncClose()
				}
			case <-queryResult.quryProcessChan:
				return 
			}	
		}()
	}
	return queryResult
}

func (pgConn *PgConn) Exec(ctx context.Context, sql string) *SimpleQuery {

	queryResult := pgConn.prepareContext(ctx)

	pgConn.frontend.Send(&proto.Query{String: sql})
	err := pgConn.flush()
	if err != nil {
		fmt.Println("exec flush err ", err)
		pgConn.asyncClose()
		queryResult.closed = true
		queryResult.err = err
		return queryResult
	}

	return queryResult
}


func (pgConn *PgConn) ExecSQL(ctx context.Context, sql string) (commandTag CommandTag, err error) {
	sq := pgConn.Exec(ctx, sql)
	commandTag, err = sq.Result()
	close(sq.quryProcessChan)
	fmt.Println("---- query --- ")
	if err != nil {
		return CommandTag{}, err
	}
	return commandTag, nil
}


type Rows struct {
	columns []proto.RowDescription
	r [][]proto.DataRow
	Len uint64
	cursor int
}

func (r *Rows) Scan(dest ...any) error {
	if len(r.columns) != len(dest) {
		return fmt.Errorf("expected %d destination arguments, got %d", len(r.columns), len(dest))
	}
	for i, val := range r.r[r.cursor] {
		if val.Value == nil {
			continue
		}
		switch d := dest[i].(type) {
		case *string:
			*d = string(val.Value)
		case *int:
			val, err := strconv.Atoi(string(val.Value))
			if err != nil {
				return fmt.Errorf("failed to scan int: %w", err)
			}
			*d = val
		case *bool:
			*d = string(val.Value) == "t"
		case *time.Time:
			dt, err := time.Parse("2006-01-02", string(val.Value))
			if err != nil {
				return fmt.Errorf("Не удалось распарсить дату")
			}
			*d = dt
		default:
			return fmt.Errorf("unsupported destination type: %T", d)
		}
	}
	return nil
}

func (r *Rows) Next() bool {
	r.cursor++
	return r.cursor < len(r.r)
}

func (pgConn *PgConn) QuerySQL(ctx context.Context, sql string, params []utils.DriverValue) (*Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, ctx.Err()
	}

	sq := pgConn.Query(ctx, sql)
	rows, err := sq.QueryResult(params)
	close(sq.quryProcessChan)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (sq *SimpleQuery) QueryResult(params []utils.DriverValue) (*Rows, error) {
	for !sq.closed && sq.err == nil {
		msg, err := sq.pgConn.receiveMessage()
			if err != nil {
				sq.err = err
				sq.closed = true
				sq.pgConn.asyncClose()
				return nil, sq.err
			}
			switch msg := msg.(type) {
			case *proto.ParseComplete:
				sq.pgConn.frontend.Send(&proto.Bind{Params: params})
				sq.pgConn.frontend.Send(&proto.Flush{})
				err := sq.pgConn.flush()
				if err != nil {
					sq.err = err
					sq.closed = true
					sq.pgConn.asyncClose()
					return nil, sq.err
				}
			case *proto.BindComplete:
				sq.pgConn.frontend.Send(&proto.Describe{})
				sq.pgConn.frontend.Send(&proto.Flush{})
				err := sq.pgConn.flush()
				if err != nil {
					sq.err = err
					sq.closed = true
					sq.pgConn.asyncClose()
					return nil, sq.err
				}
			case *proto.RowDescription:
				sq.resultKeeper.RowDescriptions = msg.Rows
				sq.pgConn.frontend.Send(&proto.Execute{})
				sq.pgConn.frontend.Send(&proto.Flush{})
				err := sq.pgConn.flush()
				if err != nil {
					sq.err = err
					sq.closed = true
					sq.pgConn.asyncClose()
					return nil, sq.err
				}
			case *proto.DataRow:
				sq.resultKeeper.DataRows = append(sq.resultKeeper.DataRows, msg.Rows)
			case *proto.CommandComplete:
				sq.pgConn.frontend.Send(&proto.Close{})
				sq.pgConn.frontend.Send(&proto.Flush{})
				err := sq.pgConn.flush()
				if err != nil {
					sq.err = err
					sq.closed = true
					return nil, sq.err
				}
				if sq.resultKeeper == nil {
					rk := &ResultKeeper{
						commandTag: CommandTag{String: string(msg.CommandTag)},
					}
					sq.resultKeeper = rk
				} else {
					sq.resultKeeper.commandTag = CommandTag{String: string(msg.CommandTag)}
				}
			case *proto.ErrorResponse:
				fmt.Println(msg.Error)
				sq.err = err
				sq.closed = true
				return nil, sq.err
			case *proto.CloseComplete:
				sq.pgConn.frontend.Send(&proto.Sync{})
				err := sq.pgConn.flush()
				if err != nil {
					sq.err = err
					sq.closed = true
					return nil, sq.err
				}
			case *proto.ReadyForQuery:
				sq.closed = true
				return &Rows{
					r: sq.resultKeeper.DataRows,
					cursor: -1,
					columns: sq.resultKeeper.RowDescriptions,
				}, nil
			}
	}
	return nil, sq.err
}

func (pgConn *PgConn) Query(ctx context.Context, query string) *SimpleQuery {

	queryResult := pgConn.prepareContext(ctx)

	pgConn.frontend.Send(&proto.Parse{Query: query})
	pgConn.frontend.Send(&proto.Flush{})
	err := pgConn.flush()
	if err != nil {
		pgConn.asyncClose()
		queryResult.closed = true
		queryResult.err = err
		return queryResult
	}

	return queryResult
}