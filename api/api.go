package api

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/validator.v8"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiPort() string
	PrometheusMonitoring() bool
	PrometheusJoinURL() string
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(elrondFacade MainApiHandler) error {
	ws := gin.Default()
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return err
	}
	registerRoutes(ws, elrondFacade)

	if elrondFacade.PrometheusMonitoring() {
		err = joinMonitoringSystem(elrondFacade.RestApiPort(), elrondFacade.PrometheusJoinURL())
		if err != nil {
			return err
		}
	}

	return ws.Run(fmt.Sprintf(":%s", elrondFacade.RestApiPort()))
}

func joinMonitoringSystem(port string, prometheusJoinURL string) error {
	req, err := http.NewRequest("POST", prometheusJoinURL, strings.NewReader(port))

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

	if elrondFacade.(MainApiHandler).PrometheusMonitoring() {
		nodeRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

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
