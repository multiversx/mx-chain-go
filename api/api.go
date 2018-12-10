package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
)

// Start will boot up the api and appropriate routes, handlers and validators
func Start(elrondFacade node.Handler) error {
	ws := gin.Default()
	ws.Use(cors.Default())

	nodeRoutes := ws.Group("/node")
	nodeRoutes.Use(middleware.WithElrondFacade(elrondFacade))
	node.Routes(nodeRoutes)

	err := ws.Run()
	return err
}
