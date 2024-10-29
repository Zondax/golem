package zprofiller

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"net/http"
	"net/http/pprof"
	pprofRuntime "runtime/pprof"
	"time"
)

const (
	defaultAddress = ":8888"
	defaultTimeOut = 300
)

type Config struct {
	ReadTimeOut  time.Duration
	WriteTimeOut time.Duration
	Logger       *logger.Logger
}

type zprofiller struct {
	router *chi.Mux
	config *Config
}

type ZProfiller interface {
	Run(addr ...string) error
}

func (c *Config) setDefaultValues() {
	if c.ReadTimeOut == 0 {
		c.ReadTimeOut = time.Duration(defaultTimeOut) * time.Millisecond
	}

	if c.WriteTimeOut == 0 {
		c.WriteTimeOut = time.Duration(defaultTimeOut) * time.Millisecond
	}

	if c.Logger == nil {
		l := logger.NewLogger()
		c.Logger = l
	}
}

func New(_ metrics.TaskMetrics, config *Config) ZProfiller {
	if config == nil {
		config = &Config{}
	}

	config.setDefaultValues()

	router := chi.NewRouter()
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	for _, profile := range pprofRuntime.Profiles() {
		router.Handle(fmt.Sprintf("/debug/pprof/%s", profile.Name()), pprof.Handler(profile.Name()))
	}

	zr := &zprofiller{
		router: router,
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

	server := &http.Server{
		Addr:         address,
		Handler:      r.router,
		ReadTimeout:  r.config.ReadTimeOut,
		WriteTimeout: r.config.WriteTimeOut,
	}

	return server.ListenAndServe()
}
