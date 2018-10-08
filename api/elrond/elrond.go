package elrond

import "github.com/gin-gonic/gin"

type Api struct{}

func (Api) AppStatus(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/appstatus"})
}

func (Api) Balance(c *gin.Context) {
	addr := c.Query("address")
	c.JSON(200, gin.H{"ok": "/balance/" + addr})
}

func (Api) CheckFreePort(c *gin.Context) {
	ip := c.Query("ipAddress")
	port := c.Query("port")
	c.JSON(200, gin.H{"ok": "/checkfreeport/" + ip + " " + port})
}

func (Api) Exit(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/exit"})
}

func (Api) GenerateKeys(c *gin.Context) {
	privateKey := c.Query("privateKey")
	c.JSON(200, gin.H{"ok": "/generatepublickeyandprivateKey/" + privateKey})
}

func (Api) GetBlockFromHash(c *gin.Context) {
	blockHash := c.Query("blockHash")
	c.JSON(200, gin.H{"ok": "/getblockfromhash/" + blockHash})
}

func (Api) GetNextPrivateKey(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/getNextPrivateKey"})
}

func (Api) GetShard(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/getprivatepublickeyshard"})
}

func (Api) GetStats(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/getStats"})
}

func (Api) GetTransactionFromHash(c *gin.Context) {
	transactionHash := c.Query("transactionHash")
	c.JSON(200, gin.H{"ok": "/gettransactionfromhash/" + transactionHash})
}

func (Api) Ping(c *gin.Context) {
	ip := c.Query("ipAddress")
	port := c.Query("port")
	c.JSON(200, gin.H{"ok": "/ping/" + ip + " " + port})

}

func (Api) Receipt(c *gin.Context) {
	transactionHash := c.Query("transactionHash")
	c.JSON(200, gin.H{"ok": "/receipt/" + transactionHash})
}

func (Api) Send(c *gin.Context) {
	adderss := c.Query("address")
	value := c.Query("value")
	c.JSON(200, gin.H{"ok": "/send/" + adderss + " " + value})

}

func (Api) SendMultipleTransactions(c *gin.Context) {
	adderss := c.Query("address")
	value := c.Query("value")
	nrTransactions := c.Query("nrTransactions")
	c.JSON(200, gin.H{"ok": "/sendMultipleTransactions/" + adderss + " " + value + " " + nrTransactions})

}

func (Api) SendMultipleTransactionsToAllShards(c *gin.Context) {
	value := c.Query("value")
	nrTransactions := c.Query("nrTransactions")
	c.JSON(200, gin.H{"ok": "/sendMultipleTransactionsToAllShards/" + value + " " + nrTransactions})
}

func (Api) ShardOfAddress(c *gin.Context) {
	addr := c.Query("address")
	c.JSON(200, gin.H{"ok": "/shardofaddress/" + addr})
}

func (Api) Start(c *gin.Context) {
	nodeName := c.Query("nodeName")
	port := c.Query("port")
	masterPeerPort := c.Query("masterPeerPort")
	masterPeerIpAddress := c.Query("masterPeerIpAddress")
	privateKey := c.Query("privateKey")
	mintValue := c.Query("mintValue")
	bootstrapType := c.Query("bootstrapType")
	c.JSON(200, gin.H{"ok": "/start/" + nodeName + " " + port + " " + masterPeerPort + " " + masterPeerIpAddress + " " + privateKey + " " + mintValue + " " + bootstrapType})
}

func (Api) Status(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/status"})
}

func (Api) Stop(c *gin.Context) {
	c.JSON(200, gin.H{"ok": "/stop"})
}
