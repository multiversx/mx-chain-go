package gin

import (
	"context"
	"net/http"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/gin/disabled"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/network"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/shared"
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

var log = logger.GetOrCreate("api/gin")

// ArgsNewWebServer holds the arguments needed to create a new instance of webServer
type ArgsNewWebServer struct {
	Facade          shared.ApiFacadeHandler
	ApiConfig       config.ApiRoutesConfig
	AntiFloodConfig config.WebServerAntifloodConfig
}

type webServer struct {
	sync.RWMutex
	facade          shared.ApiFacadeHandler
	apiConfig       config.ApiRoutesConfig
	antiFloodConfig config.WebServerAntifloodConfig
	httpServer      shared.HttpServerCloser
	cancelFunc      func()
}

// NewGinWebServerHandler returns a new instance of webServer
func NewGinWebServerHandler(args ArgsNewWebServer) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &webServer{
		facade:          args.Facade,
		antiFloodConfig: args.AntiFloodConfig,
		apiConfig:       args.ApiConfig,
	}

	return gws, nil
}

// UpdateFacade updates the main api handler by closing the old server and starting it with the new facade. Returns the
// new web server
func (ws *webServer) UpdateFacade(facade shared.ApiFacadeHandler) error {
	ws.Lock()
	ws.facade = facade
	ws.Unlock()

	closableWebServer, err := ws.CreateHttpServer()
	if err != nil {
		return err
	}

	return ws.SetHttpServer(closableWebServer)
}

// CreateHttpServer will create a new instance of http.Server and populate it with all the routes
func (ws *webServer) CreateHttpServer() (shared.HttpServerCloser, error) {
	ws.Lock()
	defer ws.Unlock()

	if ws.facade.RestApiInterface() == facade.DefaultRestPortOff {
		return disabled.NewDisabledServerClosing(), nil
	}

	var engine *gin.Engine
	if !ws.facade.RestAPIServerDebugMode() {
		gin.DefaultWriter = &ginWriter{}
		gin.DefaultErrorWriter = &ginErrorWriter{}
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	engine = gin.Default()
	engine.Use(cors.Default())
	engine.Use(middleware.WithFacade(ws.facade))

	processors, err := ws.createMiddlewareLimiters()
	if err != nil {
		return nil, err
	}

	for _, proc := range processors {
		if check.IfNil(proc) {
			continue
		}

		engine.Use(proc.MiddlewareHandlerFunc())
	}

	err = registerValidators()
	if err != nil {
		return nil, err
	}

	ws.registerRoutes(engine)

	server := &http.Server{Addr: ws.facade.RestApiInterface(), Handler: engine}
	log.Debug("creating gin web sever", "interface", ws.facade.RestApiInterface())
	wrappedServer, err := NewHttpServer(server)
	if err != nil {
		return nil, err
	}

	log.Debug("starting web server",
		"SimultaneousRequests", ws.antiFloodConfig.SimultaneousRequests,
		"SameSourceRequests", ws.antiFloodConfig.SameSourceRequests,
		"SameSourceResetIntervalInSec", ws.antiFloodConfig.SameSourceResetIntervalInSec,
	)

	return wrappedServer, nil
}

// SetHttpServer will set the inner http server to the provided one
func (ws *webServer) SetHttpServer(httpServer shared.HttpServerCloser) error {
	if check.IfNil(httpServer) {
		return errors.ErrNilHttpServer
	}

	ws.Lock()
	ws.httpServer = httpServer
	ws.Unlock()

	return nil
}

// GetHttpServer returns the closable http server
func (ws *webServer) GetHttpServer() shared.HttpServerCloser {
	ws.RLock()
	defer ws.RUnlock()

	return ws.httpServer
}

func (ws *webServer) createMiddlewareLimiters() ([]shared.MiddlewareProcessor, error) {
	sourceLimiter, err := middleware.NewSourceThrottler(ws.antiFloodConfig.SameSourceRequests)
	if err != nil {
		return nil, err
	}

	var ctx context.Context
	ctx, ws.cancelFunc = context.WithCancel(context.Background())

	go ws.sourceLimiterReset(ctx, sourceLimiter)

	globalLimiter, err := middleware.NewGlobalThrottler(ws.antiFloodConfig.SimultaneousRequests)
	if err != nil {
		return nil, err
	}

	return []shared.MiddlewareProcessor{sourceLimiter, globalLimiter}, nil
}

func (ws *webServer) sourceLimiterReset(ctx context.Context, reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(ws.antiFloodConfig.SameSourceResetIntervalInSec)
	for {
		select {
		case <-time.After(betweenResetDuration):
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

// registerRoutes has to be called under mutex protection
func (ws *webServer) registerRoutes(gws *gin.Engine) {
	routesConfig := ws.apiConfig
	nodeRoutes := gws.Group("/node")
	wrappedNodeRouter, err := wrapper.NewRouterWrapper("node", nodeRoutes, routesConfig)
	if err == nil {
		node.Routes(wrappedNodeRouter)
	}

	addressRoutes := gws.Group("/address")
	wrappedAddressRouter, err := wrapper.NewRouterWrapper("address", addressRoutes, routesConfig)
	if err == nil {
		address.Routes(wrappedAddressRouter)
	}

	networkRoutes := gws.Group("/network")
	wrappedNetworkRoutes, err := wrapper.NewRouterWrapper("network", networkRoutes, routesConfig)
	if err == nil {
		network.Routes(wrappedNetworkRoutes)
	}

	txRoutes := gws.Group("/transaction")
	wrappedTransactionRouter, err := wrapper.NewRouterWrapper("transaction", txRoutes, routesConfig)
	if err == nil {
		transaction.Routes(wrappedTransactionRouter)
	}

	vmValuesRoutes := gws.Group("/vm-values")
	wrappedVmValuesRouter, err := wrapper.NewRouterWrapper("vm-values", vmValuesRoutes, routesConfig)
	if err == nil {
		vmValues.Routes(wrappedVmValuesRouter)
	}

	validatorRoutes := gws.Group("/validator")
	wrappedValidatorsRouter, err := wrapper.NewRouterWrapper("validator", validatorRoutes, routesConfig)
	if err == nil {
		valStats.Routes(wrappedValidatorsRouter)
	}

	hardforkRoutes := gws.Group("/hardfork")
	wrappedHardforkRouter, err := wrapper.NewRouterWrapper("hardfork", hardforkRoutes, routesConfig)
	if err == nil {
		hardfork.Routes(wrappedHardforkRouter)
	}

	blockRoutes := gws.Group("/block")
	wrappedBlockRouter, err := wrapper.NewRouterWrapper("block", blockRoutes, routesConfig)
	if err == nil {
		block.Routes(wrappedBlockRouter)
	}

	if ws.facade.PprofEnabled() {
		pprof.Register(gws)
	}

	if isLogRouteEnabled(routesConfig) {
		marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
		registerLoggerWsRoute(gws, marshalizerForLogs)
	}
}

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	if ws.cancelFunc != nil {
		ws.cancelFunc()
	}

	ws.Lock()
	err := ws.httpServer.Close()
	ws.Unlock()

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ws *webServer) IsInterfaceNil() bool {
	return ws == nil
}
