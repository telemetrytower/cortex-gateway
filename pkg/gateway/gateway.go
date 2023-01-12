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
	cfg                Config
	distributorProxy   *Proxy
	queryFrontendProxy *Proxy
	rulerProxy         *Proxy
	alertmanagerProxy  *Proxy
	server             *server.Server
}

// New instantiates a new Gateway
func New(cfg Config, svr *server.Server) (*Gateway, error) {
	// Initialize reverse proxy for each upstream target service
	distributor, err := newProxy(cfg.DistributorAddress, "distributor")
	if err != nil {
		return nil, err
	}
	queryFrontend, err := newProxy(cfg.QueryFrontendAddress, "query-frontend")
	if err != nil {
		return nil, err
	}

	// Ruler Address maybe an empty string
	var ruler *Proxy
	if cfg.AlertmanagerAddress != "" {
		ruler, err = newProxy(cfg.RulerAddress, "ruler")
		if err != nil {
			return nil, err
		}
	}

	// Alertmanager Address maybe an empty string
	var alertmanager *Proxy
	if cfg.AlertmanagerAddress != "" {
		alertmanager, err = newProxy(cfg.AlertmanagerAddress, "alertmanager")
		if err != nil {
			return nil, err
		}
	}

	return &Gateway{
		cfg:                cfg,
		distributorProxy:   distributor,
		queryFrontendProxy: queryFrontend,
		rulerProxy:         ruler,
		alertmanagerProxy:  alertmanager,
		server:             svr,
	}, nil
}

// Start initializes the Gateway and starts it
func (g *Gateway) Start() {
	g.registerRoutes()
}

// RegisterRoutes binds all to be piped routes to their handlers
func (g *Gateway) registerRoutes() {
	g.server.HTTP.Path("/api/v1/push").Handler(AuthenticateTenant.Wrap(http.HandlerFunc(g.distributorProxy.Handler)))
	g.server.HTTP.PathPrefix("/prometheus").Handler(AuthenticateTenant.Wrap(http.HandlerFunc(g.queryFrontendProxy.Handler)))

	if g.rulerProxy != nil {
		g.server.HTTP.Path("/api/v1/alerts").Handler(AuthenticateTenant.Wrap(http.HandlerFunc(g.rulerProxy.Handler)))
		g.server.HTTP.PathPrefix("/api/v1/rules").Handler(AuthenticateTenant.Wrap(http.HandlerFunc(g.rulerProxy.Handler)))
	}

	if g.alertmanagerProxy != nil {
		g.server.HTTP.PathPrefix("/alertmanager").Handler(AuthenticateTenant.Wrap(http.HandlerFunc(g.alertmanagerProxy.Handler)))
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