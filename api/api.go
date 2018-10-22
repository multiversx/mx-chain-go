package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	r := gin.Default()

	node.CORSMiddleware(r)
	node.Routes(r.Group("/node"))

	viper.SetDefault("address", "127.0.0.1")
	viper.SetDefault("port", "8080")
	viper.SetConfigName("web-server")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	r.Run(viper.GetString("address") + ":" + viper.GetString("port"))
}
