package zprofiller

import (
	"github.com/zondax/golem/pkg/logger"
	"net/http"
	_ "net/http/pprof"
)

const (
	defaultAddress = ":8888"
)

type Config struct {
	Logger *logger.Logger
}

type zprofiller struct {
	config *Config
}

type ZProfiller interface {
	Run(addr ...string) error
}

func New(config *Config) ZProfiller {
	if config == nil {
		config = &Config{}
	}

	zr := &zprofiller{
		config: config,
	}

	return zr
}

func (r *zprofiller) Run(addr ...string) error {
	address := defaultAddress
	if len(addr) > 0 {
		address = addr[0]
	}

	r.config.Logger.Infof("Start profiller server at %v", address)

	return http.ListenAndServe(defaultAddress, nil)
}
