package main

import "github.com/gin-gonic/gin"

type APIer interface {
	AppStatus(*gin.Context)
	Balance(*gin.Context)
	CheckFreePort(*gin.Context)
	Exit(*gin.Context)
	GenerateKeys(*gin.Context)
	GetBlockFromHash(*gin.Context)
	GetNextPrivateKey(*gin.Context)
	GetShard(*gin.Context)
	GetStats(*gin.Context)
	GetTransactionFromHash(*gin.Context)
	Ping(*gin.Context)
	Receipt(*gin.Context)
	Send(*gin.Context)
	SendMultipleTransactions(*gin.Context)
	SendMultipleTransactionsToAllShards(*gin.Context)
	ShardOfAddress(*gin.Context)
	Start(*gin.Context)
	Status(*gin.Context)
	Stop(*gin.Context)
}
