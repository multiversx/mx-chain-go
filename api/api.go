package api

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/json"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/validator.v8"
)

var log = logger.GetOrCreate("api")

type validatorInput struct {
	Name      string
	Validator validator.Func
}

type prometheus struct {
	NodePort  string
	NetworkID string
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiPort() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	PrometheusMonitoring() bool
	PrometheusJoinURL() string
	PrometheusNetworkID() string
	IsInterfaceNil() bool
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(elrondFacade MainApiHandler) error {
	var ws *gin.Engine
	if elrondFacade.RestAPIServerDebugMode() {
		ws = gin.Default()
	} else {
		ws = gin.New()
		ws.Use(gin.Recovery())
		gin.SetMode(gin.ReleaseMode)
	}
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return err
	}

	registerRoutes(ws, elrondFacade)

	if elrondFacade.PrometheusMonitoring() {
		err = joinMonitoringSystem(elrondFacade)
		if err != nil {
			return err
		}
	}

	return ws.Run(fmt.Sprintf(":%s", elrondFacade.RestApiPort()))
}

func joinMonitoringSystem(elrondFacade MainApiHandler) error {
	prometheusJoinUrl := elrondFacade.PrometheusJoinURL()
	structToSend := prometheus{
		NodePort:  elrondFacade.RestApiPort(),
		NetworkID: elrondFacade.PrometheusNetworkID(),
	}

	jsonValue, err := json.Marshal(structToSend)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", prometheusJoinUrl, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = resp.Body.Close()
	return err
}

func registerRoutes(ws *gin.Engine, elrondFacade middleware.ElrondHandler) {
	nodeRoutes := ws.Group("/node")
	nodeRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	node.Routes(nodeRoutes)

	addressRoutes := ws.Group("/address")
	addressRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	address.Routes(addressRoutes)

	txRoutes := ws.Group("/transaction")
	txRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	transaction.Routes(txRoutes)

	vmValuesRoutes := ws.Group("/vm-values")
	vmValuesRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	vmValues.Routes(vmValuesRoutes)

	apiHandler, ok := elrondFacade.(MainApiHandler)
	if ok && apiHandler.PrometheusMonitoring() {
		nodeRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	if apiHandler.PprofEnabled() {
		pprof.Register(ws)
	}

	marshalizerForLogs := &marshal.ProtobufMarshalizer{}
	registerLoggerWsRoute(ws, marshalizerForLogs)
}

func registerValidators() error {
	validators := []validatorInput{
		{Name: "skValidator", Validator: skValidator},
	}
	for _, validatorFunc := range validators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			err := v.RegisterValidation(validatorFunc.Name, validatorFunc.Validator)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}
