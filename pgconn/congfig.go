package pgconn

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TiJon8/cpol/proto"
)


type ConnectionConfig struct {
	Host           string // host (e.g. localhost) or absolute path to unix domain socket directory (e.g. /private/tmp)
	Port           uint16
	User           string
	Password       string
	Database       string
	ProtocolVersion string

	ConnectTimeout time.Duration
	DialFunc       DialFunc

	createdByParsingConfig bool

	RuntimeParams map[string]string
	FrontendBuilder FrontendBuilderFunc

	LookupFunc LookupFunc
}


func ParseConfigMap(configMap map[string]string) (*ConnectionConfig, error) {
	connectSettings := make(map[string]string)

	connectParams := map[string]struct{}{
		"host": {},
		"port": {},
		"user": {},
		"password": {},
		"database": {},
	}
	for k, v := range configMap {
		if _, exists := connectParams[k]; exists {
			connectSettings[k] = v
		}
	}
	connString := fmt.Sprintf(
		"//%s:%s@%s:%s/%s",
		connectSettings["user"],
		connectSettings["password"],
		connectSettings["host"],
		connectSettings["port"],
		connectSettings["database"],
	)
	parsedConnString, err := parseUrlSettings(connString)
	if err != nil {
		return nil, err
	}
	configSettings := mergeSettings(parsedConnString, configMap)
	return CollectConfig(configSettings)
}


func CollectConfig(parsedConfigSettings map[string]string) (*ConnectionConfig, error) {
	defaultSettings := defaultSettings()
	envSettings := envSettings()

	settings := mergeSettings(defaultSettings, envSettings, parsedConfigSettings)
	fmt.Println(settings)
	config := &ConnectionConfig{
		Database: settings["database"],
		User: settings["user"],
		Password: settings["password"],
		Host: settings["host"],
		createdByParsingConfig: true,
		RuntimeParams: make(map[string]string),
		FrontendBuilder: func(r io.Reader, w io.Writer) *proto.Frontend {
			return proto.NewFrontend(r, w)
		},
	}

	if connectTimeoutSetting, exists := settings["connect_timeout"]; exists {
		connectTimeout, err := parseConnectTimeoutSetting(connectTimeoutSetting)
		if err != nil {
			return nil, err
		}
		config.ConnectTimeout = connectTimeout
		config.DialFunc = TimeoutDialFunc(connectTimeout)
	} else {
		defaultDialer := makeDefaultDialer()
		config.DialFunc = defaultDialer.DialContext
	}

	config.LookupFunc = makeDefaultResolver().LookupHost

	nonRuntimeParams := map[string]struct{}{
		"host":                 {},
		"port":                 {},
		"database":             {},
		"user":                 {},
		"password":             {},
		"protocol_version": 	 {},
	}

	for k, v := range settings {
		if _, exists := nonRuntimeParams[k]; exists {
			continue
		}
		config.RuntimeParams[k] = v
	}

	parsedPort, err := parsePort(settings["port"])
	if err != nil {
		return nil, err
	}
	config.Port = parsedPort

	config.ProtocolVersion = settings["protocol_version"]
	if config.ProtocolVersion == "" {
		config.ProtocolVersion = "3.2"
	}
	fmt.Println(config)
	return config, nil
}

func TimeoutDialFunc(d time.Duration) DialFunc {
	dialer := makeDefaultDialer()
	dialer.Timeout = d
	return dialer.DialContext
}
func makeDefaultDialer() *net.Dialer {
	return &net.Dialer{}
}

func makeDefaultResolver() *net.Resolver {
	return net.DefaultResolver
}

func parsePort(p string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(p, 10, 16)
	if err != nil {
		return 0, err
	}
	if parsedPort < 1 || parsedPort > math.MaxUint16 {
		return 0, errors.New("port number is out of range")
	}
	return uint16(parsedPort), nil
}

func parseConnectTimeoutSetting(connectTimeout string) (time.Duration, error) {
	timeout, err := strconv.ParseInt(connectTimeout, 10, 64)
	if err != nil {
		return 0, err
	}
	if timeout < 0 {
		return 0, errors.New("negative timeout")
	}
	return time.Duration(timeout) * time.Second, nil
}

func mergeSettings(settings ...map[string]string) map[string]string {
	collectSettings := make(map[string]string)

	for _, v := range settings {
		maps.Copy(collectSettings, v)
	}
	return collectSettings
}


func parseUrlSettings(connString string) (map[string]string, error) {
	settings := make(map[string]string)

	parsedUrl, err := url.Parse(connString)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(parsedUrl) // debug
	if parsedUrl.User != nil {
		if u := parsedUrl.User.Username(); u != "" {
			settings["user"] = u
		}
		if password, exists := parsedUrl.User.Password(); exists {
			settings["password"] = password
		}
	}

	h, p, err := net.SplitHostPort(parsedUrl.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to split host:port in '%s', err: %w", parsedUrl.Host, err)
	}
	if h != "" {
		settings["host"] = h
	}
	if p != "" {
		settings["port"] = p
	}

	database := strings.TrimLeft(parsedUrl.Path, "/")
	if database != "" {
		settings["database"] = database
	}

	return settings, nil
}


func parseProtocolVersion(s string) (uint32, error) {
	switch s {
	case "3.2", "latest":
		return proto.ProtocolVersionLatest, nil
	default:
		return 0, fmt.Errorf("invalid proto version: %q", s)
	}
}
