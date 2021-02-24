package ginwebserver

import (
	"context"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/api/factory"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/network"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	valStats "github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/factory")

type ginWebServerHandler struct {
	facade          factory.MainApiHandler
	apiConfig       config.ApiRoutesConfig
	antiFloodConfig config.WebServerAntifloodConfig
	httpServer      factory.HttpServerClosingHandler
	ctx             context.Context
	cancelFunc      func()
}

// ArgsNewGinWebServerHandler holds the arguments needed to create a new instance of ginWebServerHandler
type ArgsNewGinWebServerHandler struct {
	Facade          factory.MainApiHandler
	ApiConfig       config.ApiRoutesConfig
	AntiFloodConfig config.WebServerAntifloodConfig
}

// NewGinWebServerHandler returns a new instance of ginWebServerHandler
func NewGinWebServerHandler(args ArgsNewGinWebServerHandler) (*ginWebServerHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &ginWebServerHandler{
		facade:          args.Facade,
		antiFloodConfig: args.AntiFloodConfig,
		apiConfig:       args.ApiConfig,
	}

	gws.ctx, gws.cancelFunc = context.WithCancel(context.Background())

	return gws, nil
}

// UpdateFacade updates the main api handler by closing the old server and starting it with the new facade. Returns the
// new web server
func (gws *ginWebServerHandler) UpdateFacade(facade factory.MainApiHandler) (factory.HttpServerClosingHandler, error) {
	if gws.httpServer != nil {
		err := gws.httpServer.Close()
		if err != nil {
			return nil, err
		}
	}

	gws.facade = facade
	webServer, err := gws.CreateHttpServer()
	if err != nil {
		return nil, err
	}

	gws.httpServer = webServer

	return webServer, nil
}

// CreateHttpServer will create a new instance of http.Server and populate it with all the routes
func (gws *ginWebServerHandler) CreateHttpServer() (factory.HttpServerClosingHandler, error) {
	if gws.facade.RestApiInterface() == facade.DefaultRestPortOff {
		return NewDisabledServerClosing(), nil
	}

	var ws *gin.Engine
	if !gws.facade.RestAPIServerDebugMode() {
		gin.DefaultWriter = &ginWriter{}
		gin.DefaultErrorWriter = &ginErrorWriter{}
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	ws = gin.Default()
	ws.Use(cors.Default())
	ws.Use(middleware.WithFacade(gws.facade))

	processors, err := gws.createMiddlewareLimiters()
	if err != nil {
		return nil, err
	}

	for _, proc := range processors {
		if check.IfNil(proc) {
			continue
		}

		ws.Use(proc.MiddlewareHandlerFunc())
	}

	err = registerValidators()
	if err != nil {
		return nil, err
	}

	gws.registerRoutes(ws)

	server := &http.Server{Addr: gws.facade.RestApiInterface(), Handler: ws}
	log.Debug("creating gin web sever", "interface", gws.facade.RestApiInterface())
	wrappedServer, err := NewHttpServer(server)
	if err != nil {
		return nil, err
	}

	log.Debug("starting web server",
		"SimultaneousRequests", gws.antiFloodConfig.SimultaneousRequests,
		"SameSourceRequests", gws.antiFloodConfig.SameSourceRequests,
		"SameSourceResetIntervalInSec", gws.antiFloodConfig.SameSourceResetIntervalInSec,
	)

	gws.httpServer = wrappedServer

	return wrappedServer, nil
}

func (gws *ginWebServerHandler) createMiddlewareLimiters() ([]factory.MiddlewareProcessor, error) {
	sourceLimiter, err := middleware.NewSourceThrottler(gws.antiFloodConfig.SameSourceRequests)
	if err != nil {
		return nil, err
	}
	go gws.sourceLimiterReset(sourceLimiter)

	globalLimiter, err := middleware.NewGlobalThrottler(gws.antiFloodConfig.SimultaneousRequests)
	if err != nil {
		return nil, err
	}

	return []factory.MiddlewareProcessor{sourceLimiter, globalLimiter}, nil
}

func (gws *ginWebServerHandler) sourceLimiterReset(reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(gws.antiFloodConfig.SameSourceResetIntervalInSec)
	for {
		select {
		case <-time.After(betweenResetDuration):
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-gws.ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

func (gws *ginWebServerHandler) registerRoutes(ws *gin.Engine) {
	routesConfig := gws.apiConfig
	nodeRoutes := ws.Group("/node")
	wrappedNodeRouter, err := wrapper.NewRouterWrapper("node", nodeRoutes, routesConfig)
	if err == nil {
		node.Routes(wrappedNodeRouter)
	}

	addressRoutes := ws.Group("/address")
	wrappedAddressRouter, err := wrapper.NewRouterWrapper("address", addressRoutes, routesConfig)
	if err == nil {
		address.Routes(wrappedAddressRouter)
	}

	networkRoutes := ws.Group("/network")
	wrappedNetworkRoutes, err := wrapper.NewRouterWrapper("network", networkRoutes, routesConfig)
	if err == nil {
		network.Routes(wrappedNetworkRoutes)
	}

	txRoutes := ws.Group("/transaction")
	wrappedTransactionRouter, err := wrapper.NewRouterWrapper("transaction", txRoutes, routesConfig)
	if err == nil {
		transaction.Routes(wrappedTransactionRouter)
	}

	vmValuesRoutes := ws.Group("/vm-values")
	wrappedVmValuesRouter, err := wrapper.NewRouterWrapper("vm-values", vmValuesRoutes, routesConfig)
	if err == nil {
		vmValues.Routes(wrappedVmValuesRouter)
	}

	validatorRoutes := ws.Group("/validator")
	wrappedValidatorsRouter, err := wrapper.NewRouterWrapper("validator", validatorRoutes, routesConfig)
	if err == nil {
		valStats.Routes(wrappedValidatorsRouter)
	}

	hardforkRoutes := ws.Group("/hardfork")
	wrappedHardforkRouter, err := wrapper.NewRouterWrapper("hardfork", hardforkRoutes, routesConfig)
	if err == nil {
		hardfork.Routes(wrappedHardforkRouter)
	}

	blockRoutes := ws.Group("/block")
	wrappedBlockRouter, err := wrapper.NewRouterWrapper("block", blockRoutes, routesConfig)
	if err == nil {
		block.Routes(wrappedBlockRouter)
	}

	if gws.facade.PprofEnabled() {
		pprof.Register(ws)
	}

	if isLogRouteEnabled(routesConfig) {
		marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
		registerLoggerWsRoute(ws, marshalizerForLogs)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (gws *ginWebServerHandler) IsInterfaceNil() bool {
	return gws == nil
}
