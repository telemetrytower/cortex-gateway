package gateway

import (
	"net/http"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/weaveworks/common/server"
)

// Gateway hosts a reverse proxy for each upstream cortex service we'd like to tunnel after successful authentication
type Gateway struct {
	cfg     Config
	proxies map[string]*Proxy
	server  *server.Server
}

// New instantiates a new Gateway
func New(cfg Config, svr *server.Server) (*Gateway, error) {
	// init proxies
	proxies := map[string]*Proxy{}
	for targetName, target := range cfg.Targets {
		proxy, err := newProxy(target, targetName)
		if err != nil {
			return nil, err
		}

		proxies[targetName] = proxy
	}

	return &Gateway{
		cfg:     cfg,
		proxies: proxies,
		server:  svr,
	}, nil
}

// Start initializes the Gateway and starts it
func (g *Gateway) Start() {
	g.registerRoutes()
}

// RegisterRoutes binds all to be piped routes to their handlers
func (g *Gateway) registerRoutes() {
	for _, route := range g.cfg.Routes {
		proxy, ok := g.proxies[route.Target]
		if !ok {
			continue
		}

		if route.Prefix {
			g.server.HTTP.PathPrefix(route.Path).Handler(AuthenticateTenant.Wrap(http.HandlerFunc(proxy.Handler)))
		} else {
			g.server.HTTP.Path(route.Path).Handler(AuthenticateTenant.Wrap(http.HandlerFunc(proxy.Handler)))
		}
	}
	g.server.HTTP.Path("/health").HandlerFunc(g.healthCheck)
	g.server.HTTP.PathPrefix("/").HandlerFunc(g.notFoundHandler)
}

func (g *Gateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Ok"))
}

func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.With(util_log.WithContext(r.Context(), util_log.Logger), "ip_address", r.RemoteAddr)
	level.Info(logger).Log("msg", "no request handler defined for this route", "route", r.RequestURI)
	w.WriteHeader(404)
	w.Write([]byte("404 - Resource not found"))
}
