package gin

import (
	"fmt"
	"net/http"
	"reflect"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/websocket"
	"gopkg.in/go-playground/validator.v8"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
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

func checkArgs(args ArgsNewWebServer) error {
	errHandler := func(details string) error {
		return fmt.Errorf("%w: %s", apiErrors.ErrCannotCreateGinWebServer, details)
	}

	if check.IfNil(args.Facade) {
		return errHandler("nil facade")
	}

	return nil
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

func registerValidators() error {
	validators := []validatorInput{
		{
			Name:      "skValidator",
			Validator: skValidator,
		},
	}
	for _, validatorFunc := range validators {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if ok {
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
