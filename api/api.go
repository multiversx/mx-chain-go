package api

import (
	"net/http"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/network"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/proof"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	valStats "github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var log = logger.GetOrCreate("api")

// MiddlewareProcessor defines a processor used internally by the web server when processing requests
type MiddlewareProcessor interface {
	MiddlewareHandlerFunc() gin.HandlerFunc
	IsInterfaceNil() bool
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

// RegisterRoutes will register all routes available on the web server
func RegisterRoutes(ws *gin.Engine, routesConfig config.ApiRoutesConfig, elrondFacade middleware.Handler) {
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

	proofRoutes := ws.Group("/proof")
	wrappedProofRouter, err := wrapper.NewRouterWrapper("proof", proofRoutes, routesConfig)
	if err == nil {
		proof.Routes(wrappedProofRouter)
	}

	apiHandler, ok := elrondFacade.(MainApiHandler)
	if ok && apiHandler.PprofEnabled() {
		pprof.Register(ws)
	}

	if isLogRouteEnabled(routesConfig) {
		marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
		registerLoggerWsRoute(ws, marshalizerForLogs)
	}
}

func isLogRouteEnabled(routesConfig config.ApiRoutesConfig) bool {
	logConfig, ok := routesConfig.APIPackages["log"]
	if !ok {
		return false
	}

	for _, cfg := range logConfig.Routes {
		if cfg.Name == "/log" && cfg.Open {
			return true
		}
	}

	return false
}

func registerLoggerWsRoute(ws *gin.Engine, marshalizer marshal.Marshalizer) {
	upgrader := websocket.Upgrader{}

	ws.GET("/log", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls, err := logs.NewLogSender(marshalizer, conn, log)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls.StartSendingBlocking()
	})
}
