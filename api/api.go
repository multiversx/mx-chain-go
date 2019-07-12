package api

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/node"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/validator.v8"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

type prometheusDetails struct {
	NodePort  string
	NetworkID string
}

// MainApiHandler interface defines methods that can be used from `elrondFacade` context variable
type MainApiHandler interface {
	RestApiPort() string
	PrometheusMonitoring() bool
	PrometheusJoinURL() string
	PrometheusNetworkID() string
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
		err = joinMonitoringSystem(elrondFacade)
		if err != nil {
			return err
		}
	}

	return ws.Run(fmt.Sprintf(":%s", elrondFacade.RestApiPort()))
}

func joinMonitoringSystem(elrondFacade MainApiHandler) error {
	prometheusJoinUrl := elrondFacade.PrometheusJoinURL()
	structToSend := prometheusDetails{
		NodePort:  elrondFacade.RestApiPort(),
		NetworkID: elrondFacade.PrometheusNetworkID(),
	}
	jsonValue, err := json.Marshal(structToSend)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", prometheusJoinUrl, bytes.NewBuffer(jsonValue))

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

	apiHandler, ok := elrondFacade.(MainApiHandler)
	if ok && apiHandler.PrometheusMonitoring() {
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
