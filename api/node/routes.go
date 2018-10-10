package node

import "github.com/gin-gonic/gin"

var api Api

func Routes(router *gin.RouterGroup) {
	router.GET("/appstatus", api.AppStatus)
	router.GET("/balance", api.Balance)
	router.GET("/checkfreeport", api.CheckFreePort)
	router.GET("/exit", api.Exit)
	router.GET("/generatepublickeyandprivateKey", api.GenerateKeys)
	router.GET("/getblockfromhash", api.GetBlockFromHash)
	router.GET("/getNextPrivateKey", api.GetNextPrivateKey)
	router.GET("/getprivatepublickeyshard", api.GetShard)
	router.GET("/getStats", api.GetStats)
	router.GET("/gettransactionfromhash", api.GetTransactionFromHash)
	router.GET("/ping", api.Ping)
	router.GET("/receipt", api.Receipt)
	router.GET("/send", api.Send)
	router.GET("/sendMultipleTransactions", api.SendMultipleTransactions)
	router.GET("/sendMultipleTransactionsToAllShards", api.SendMultipleTransactionsToAllShards)
	router.GET("/shardofaddress", api.ShardOfAddress)
	router.GET("/start", api.Start)
	router.GET("/status", api.Status)
	router.GET("/stop", api.Stop)
}
