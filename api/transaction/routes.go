package transaction

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
}

// GenerateTx represents the structure on which user input for generating a new transaction will validate against
type GenerateTx struct {
	SecretKey string `form:"sk" json:"sk" binding:"skValidator"`
}

// Routes function defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/", GenerateTransaction)
	router.GET("/:txhash", GetTransaction)
}

// GenerateTransaction generates a new transaction given an sk and additional data
func GenerateTransaction(c *gin.Context) {
	var gtx = GenerateTx{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Validation error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {
	txhash := c.Param("txhash")
	c.JSON(http.StatusOK, gin.H{"message": txhash})
}
