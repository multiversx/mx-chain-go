package api

import (
	"net/http"
	"reflect"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	valStats "github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/websocket"
	"gopkg.in/go-playground/validator.v8"
)

var log = logger.GetOrCreate("api")

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

type ginWriter struct {
}

func (gv *ginWriter) Write(p []byte) (n int, err error) {
	log.Debug("gin server", "message", string(p))

	return len(p), nil
}

type ginErrorWriter struct {
}

func (gev *ginErrorWriter) Write(p []byte) (n int, err error) {
	log.Debug("gin server", "error", string(p))

	return len(p), nil
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(elrondFacade MainApiHandler) error {
	var ws *gin.Engine
	if !elrondFacade.RestAPIServerDebugMode() {
		gin.DefaultWriter = &ginWriter{}
		gin.DefaultErrorWriter = &ginErrorWriter{}
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	ws = gin.Default()
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return err
	}

	registerRoutes(ws, elrondFacade)

	return ws.Run(elrondFacade.RestApiInterface())
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

	validatorRoutes := ws.Group("/validator")
	validatorRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	valStats.Routes(validatorRoutes)

	apiHandler, ok := elrondFacade.(MainApiHandler)
	if ok && apiHandler.PprofEnabled() {
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
