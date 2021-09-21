package gin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/gin")

// ArgsNewWebServer holds the arguments needed to create a new instance of webServer
type ArgsNewWebServer struct {
	Facade          shared.FacadeHandler
	ApiConfig       config.ApiRoutesConfig
	AntiFloodConfig config.WebServerAntifloodConfig
}

type webServer struct {
	sync.RWMutex
	facade          shared.FacadeHandler
	apiConfig       config.ApiRoutesConfig
	antiFloodConfig config.WebServerAntifloodConfig
	httpServer      shared.HttpServerCloser
	groups          map[string]shared.GroupHandler
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
func (ws *webServer) UpdateFacade(facade shared.FacadeHandler) error {
	ws.Lock()
	defer ws.Unlock()

	ws.facade = facade

	for groupName, groupHandler := range ws.groups {
		log.Debug("upgrading facade for gin API group", "group name", groupName)
		err := groupHandler.UpdateFacade(facade)
		if err != nil {
			log.Error("cannot update facade for gin API group", "group name", groupName, "error", err)
		}
	}

	return nil
}

// CreateHttpServer will create a new instance of http.Server and populate it with all the routes
func (ws *webServer) StartHttpServer() error {
	ws.Lock()
	defer ws.Unlock()

	if ws.facade.RestApiInterface() == facade.DefaultRestPortOff {
		return nil
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

	processors, err := ws.createMiddlewareLimiters()
	if err != nil {
		return err
	}

	for _, proc := range processors {
		if check.IfNil(proc) {
			continue
		}

		engine.Use(proc.MiddlewareHandlerFunc())
	}

	err = registerValidators()
	if err != nil {
		return err
	}

	err = ws.createGroups()
	if err != nil {
		return err
	}

	ws.registerRoutes(engine)

	server := &http.Server{Addr: ws.facade.RestApiInterface(), Handler: engine}
	log.Debug("creating gin web sever", "interface", ws.facade.RestApiInterface())
	ws.httpServer, err = NewHttpServer(server)
	if err != nil {
		return err
	}

	log.Debug("starting web server",
		"SimultaneousRequests", ws.antiFloodConfig.SimultaneousRequests,
		"SameSourceRequests", ws.antiFloodConfig.SameSourceRequests,
		"SameSourceResetIntervalInSec", ws.antiFloodConfig.SameSourceResetIntervalInSec,
	)

	go ws.httpServer.Start()

	return nil
}

func (ws *webServer) createGroups() error {
	groupsMap := make(map[string]shared.GroupHandler)
	addressGroup, err := groups.NewAddressGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["address"] = addressGroup

	blockGroup, err := groups.NewBlockGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["block"] = blockGroup

	hardforkGroup, err := groups.NewHardforkGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["hardfork"] = hardforkGroup

	networkGroup, err := groups.NewNetworkGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["network"] = networkGroup

	nodeGroup, err := groups.NewNodeGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["node"] = nodeGroup

	proofGroup, err := groups.NewProofGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["proof"] = proofGroup

	transactionGroup, err := groups.NewTransactionGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["transaction"] = transactionGroup

	validatorGroup, err := groups.NewValidatorGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["validator"] = validatorGroup

	vmValuesGroup, err := groups.NewVmValuesGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["vm-values"] = vmValuesGroup

	ws.groups = groupsMap

	return nil
}

func (ws *webServer) registerRoutes(ginRouter *gin.Engine) {
	for groupName, groupHandler := range ws.groups {
		log.Debug("registering gin API group", "group name", groupName)
		ginGroup := ginRouter.Group(fmt.Sprintf("/%s", groupName))
		groupHandler.RegisterRoutes(ginGroup, ws.apiConfig)
	}

	if isLogRouteEnabled(ws.apiConfig) {
		marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
		registerLoggerWsRoute(ginRouter, marshalizerForLogs)
	}
}

func (ws *webServer) createMiddlewareLimiters() ([]shared.MiddlewareProcessor, error) {
	middlewares := make([]shared.MiddlewareProcessor, 0)

	if ws.apiConfig.Logging.LoggingEnabled {
		responseLoggerMiddleware := middleware.NewResponseLoggerMiddleware(time.Duration(ws.apiConfig.Logging.ThresholdInMicroSeconds) * time.Microsecond)
		middlewares = append(middlewares, responseLoggerMiddleware)
	}

	sourceLimiter, err := middleware.NewSourceThrottler(ws.antiFloodConfig.SameSourceRequests)
	if err != nil {
		return nil, err
	}

	var ctx context.Context
	ctx, ws.cancelFunc = context.WithCancel(context.Background())

	go ws.sourceLimiterReset(ctx, sourceLimiter)

	middlewares = append(middlewares, sourceLimiter)

	globalLimiter, err := middleware.NewGlobalThrottler(ws.antiFloodConfig.SimultaneousRequests)
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, globalLimiter)

	return middlewares, nil
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

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	if ws.cancelFunc != nil {
		ws.cancelFunc()
	}

	ws.Lock()
	err := ws.httpServer.Close()
	ws.Unlock()

	if err != nil {
		err = fmt.Errorf("%w while closing the http server in gin/webServer", err)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ws *webServer) IsInterfaceNil() bool {
	return ws == nil
}
