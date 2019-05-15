package api

import (
	"reflect"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/block"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/address"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(elrondFacade middleware.ElrondHandler) error {
	ws := gin.Default()
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return err
	}
	registerRoutes(ws, elrondFacade)

	return ws.Run()
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

	blockRoutes := ws.Group("/block")
	blockRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	block.Routes(blockRoutes)

	blocksRoutes := ws.Group("/blocks")
	blocksRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	block.RoutesForBlockLists(blocksRoutes)
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
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	return true
}
