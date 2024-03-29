package main

import (
	"flag"
	"net/http"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/server"
	"github.com/weaveworks/common/tracing"
	"google.golang.org/grpc"

	"github.com/telemetrytower/cortex-gateway/pkg/gateway"
)

var cfgPath string

func main() {
	flag.StringVar(&cfgPath, "config.file", "cortex-gateway.yaml", "Cortex gateway configuration file path")

	operationNameFunc := nethttp.OperationNameFunc(func(r *http.Request) string {
		return r.URL.RequestURI()
	})

	var (
		serverCfg = server.Config{
			MetricsNamespace: "cortex_gateway",
			HTTPMiddleware: []middleware.Interface{
				middleware.Func(func(handler http.Handler) http.Handler {
					return nethttp.Middleware(opentracing.GlobalTracer(), handler, operationNameFunc)
				}),
			},
			GRPCMiddleware: []grpc.UnaryServerInterceptor{
				middleware.ServerUserHeaderInterceptor,
			},
		}
		gatewayCfg gateway.Config
	)

	flagext.RegisterFlags(&serverCfg)
	flag.Parse()

	util_log.InitLogger(&serverCfg)

	// Must be done after initializing the logger, otherwise no log message is printed
	err := gatewayCfg.Load(cfgPath)
	util_log.CheckFatal("validating gateway config", err)

	// Setting the environment variable JAEGER_AGENT_HOST enables tracing
	trace, err := tracing.NewFromEnv("cortex-gateway")
	util_log.CheckFatal("initializing tracing", err)
	defer trace.Close()

	svr, err := server.New(serverCfg)
	util_log.CheckFatal("initializing server", err)
	defer svr.Shutdown()

	// Setup proxy and register routes
	gateway, err := gateway.New(gatewayCfg, svr)
	util_log.CheckFatal("initializing gateway", err)
	gateway.Start()

	svr.Run()
}
