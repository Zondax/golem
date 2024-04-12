package zprofiller

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"net/http"
	"net/http/pprof"
	pprofRuntime "runtime/pprof"
)

const (
	defaultAddress = ":8888"
)

type Config struct {
	Logger *logger.Logger
}

type zprofiller struct {
	router *chi.Mux
	config *Config
}

type ZProfiller interface {
	Run(addr ...string) error
}

func New(_ metrics.TaskMetrics, config *Config) ZProfiller {
	if config == nil {
		config = &Config{}
	}

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
		Addr:    address,
		Handler: r.router,
	}

	return server.ListenAndServe()
}
