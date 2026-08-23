package pgpool

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/TiJon8/cpol/pgconn"
	"github.com/TiJon8/cpol/utils"
)

const (
	defaultMaxConnections = uint32(4)
	defaultMinConnections = uint32(4)
	defaultPreOpenConnections = uint32(0)
	defaultMaxConnIdleTimeout = time.Minute*2
)

type PoolConfig struct {
	ConnectionConfig *pgconn.ConnectionConfig

	createdByParsingConfig bool

	MaxConnections uint32
	MinConnections uint32
	MaxConnIdleTimeout time.Duration
	PreOpenConnections uint32
}

type Pool struct {
	store []*ConnResource
	config *PoolConfig
	maxConns              uint32
	minConns              uint32
	maxConnIdleTimeout       time.Duration

	oppenedConnections uint32
	
	// closeOnce sync.Once
	// closeChan chan struct{}

	mutex sync.Mutex
	waitQueue map[uint64]chan *ConnResource
	nextT uint64
	ownerT uint64

	expiredConnections []*ConnResource
	isClosing bool
	wgFlight sync.WaitGroup
}


func InitWithConfig(ctx context.Context, config *PoolConfig) (*Pool, error) {
	if !config.createdByParsingConfig {
		panic("not validated by config")
	}

	var pc []*ConnResource
	var oppenedConns uint32
	fmt.Println(config)
	// при инизиализации массив соединений должен быть пустым или предоткрытым
	if config.PreOpenConnections != 0 && config.PreOpenConnections <= config.MaxConnections {
		oppenedConns = config.PreOpenConnections
		for c := range oppenedConns {
			fmt.Println("range connections ----------------------", c, oppenedConns)
			pgConn, err := pgconn.ConnectConfig(ctx, config.ConnectionConfig)
			if err != nil {
				return nil, err
			}
			pc = append(pc, &ConnResource{
				conn: pgConn,
				lastUsed: time.Now(),
			})
		}
	} else {
		oppenedConns = 0
	}

	var oT uint64 = 0
	p := &Pool{
		store: pc,
		oppenedConnections: oppenedConns,
		waitQueue: make(map[uint64]chan *ConnResource),
		ownerT: oT,
		nextT: oT+1,
		isClosing: false,

		config: config,
		minConns:              config.MinConnections,
		maxConns:              config.MaxConnections,
		maxConnIdleTimeout:       config.MaxConnIdleTimeout,
	}
	go p.connectionJanitor()

	return p, nil
}


func ParseConfigMap(configMap map[string]string) (*PoolConfig, error) {
	connectionConfig, err := pgconn.ParseConfigMap(configMap)
	if err != nil {
		return nil, err
	}

	config := &PoolConfig{
		ConnectionConfig: connectionConfig,
		createdByParsingConfig: true,
	}

	if value, exists := connectionConfig.RuntimeParams["pool_max_connections"]; exists {
		delete(connectionConfig.RuntimeParams, "pool_max_connections")
		numConns, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		config.MaxConnections = uint32(numConns)
	} else {
		config.MaxConnections = defaultMaxConnections
		if numCPU := uint32(runtime.NumCPU()); numCPU > config.MaxConnections {
			config.MaxConnections = numCPU
		} 
	}

	if value, exists := connectionConfig.RuntimeParams["pool_min_connections"]; exists {
		delete(connectionConfig.RuntimeParams, "pool_min_connections")
		minConns, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		config.MinConnections = uint32(minConns)
	} else {
		config.MinConnections = defaultMinConnections
	}

	if value, exists := connectionConfig.RuntimeParams["pre_open_connections"]; exists {
		delete(connectionConfig.RuntimeParams, "pre_open_connections")
		preOpenConns, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		config.PreOpenConnections = uint32(preOpenConns)
	} else {
		config.PreOpenConnections = defaultPreOpenConnections
	}

	if value, exists := connectionConfig.RuntimeParams["pool_max_conn_idle_timeout"]; exists {
		delete(connectionConfig.RuntimeParams, "pool_max_conn_idle_timeout")
		maxIdleTimeout, err := time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
		config.MaxConnIdleTimeout = maxIdleTimeout
	} else {
		config.MaxConnIdleTimeout = defaultMaxConnIdleTimeout
	}

	return config, nil
}

func (p *Pool) Close(ctx context.Context) []error {
	p.mutex.Lock()
	if p.isClosing {
		p.mutex.Unlock()
		return nil
	}
	p.isClosing = true

	for t, pConn := range p.waitQueue {
		close(pConn)
		delete(p.waitQueue, t)
	}

	for _, expiredConn := range p.expiredConnections {
		p.store = append(p.store, expiredConn)
	}
	p.expiredConnections = nil
	p.mutex.Unlock()

	p.wgFlight.Wait()

	p.mutex.Lock()
	var closingErrors []error
	for _, pConn := range p.store {
		fmt.Println("Closing connection by calling Close", pConn)
		if err := pConn.conn.Close(ctx); err != nil {
			closingErrors = append(closingErrors, err)
		}
		p.oppenedConnections--
	}
	p.store = nil
	p.mutex.Unlock()

	if len(closingErrors) > 0 {
		return closingErrors
	}
	fmt.Println("Pool has closed")
	return nil
}




func (p *Pool) Ping(ctx context.Context) error {
	c, err := p.Get(ctx)
	if err != nil {
		return err
	}
	defer p.Put(c)
	return c.conn.Ping(ctx)
}

func (p *Pool) Exec(ctx context.Context, sql string, parameters... any) (pgconn.CommandTag, error) {
	c, err := p.Get(ctx)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	defer p.Put(c)
	return c.conn.ExecSQL(ctx, sql)
}

func (p *Pool) Query(ctx context.Context, sql string, params ...any) (*pgconn.Rows, error) {
	sp := utils.Serialize(params)
	c, err := p.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer p.Put(c)
	return c.conn.QuerySQL(ctx, sql, sp)
}
