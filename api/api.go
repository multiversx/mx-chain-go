package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/api/elrond"
	"github.com/gin-gonic/gin"
)

func main() {
	r := SetupRouter()
	r.Run(":8080")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := elrond.Api{}

	node := r.Group("/node")
	{
		node.GET("/appstatus", api.AppStatus)
		node.GET("/balance", api.Balance)
		node.GET("/checkfreeport", api.CheckFreePort)
		node.GET("/exit", api.Exit)
		node.GET("/generatepublickeyandprivateKey", api.GenerateKeys)
		node.GET("/getblockfromhash", api.GetBlockFromHash)
		node.GET("/getNextPrivateKey", api.GetNextPrivateKey)
		node.GET("/getprivatepublickeyshard", api.GetShard)
		node.GET("/getStats", api.GetStats)
		node.GET("/gettransactionfromhash", api.GetTransactionFromHash)
		node.GET("/ping", api.Ping)
		node.GET("/receipt", api.Receipt)
		node.GET("/send", api.Send)
		node.GET("/sendMultipleTransactions", api.SendMultipleTransactions)
		node.GET("/sendMultipleTransactionsToAllShards", api.SendMultipleTransactionsToAllShards)
		node.GET("/shardofaddress", api.ShardOfAddress)
		node.GET("/start", api.Start)
		node.GET("/status", api.Status)
		node.GET("/stop", api.Stop)
	}

	return r
}
