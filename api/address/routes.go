package address

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {

}

// Routes function defines address related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/:address", GetAddress)
	router.GET("/:address/balance", GetBalance)
}

func GetAddress(c *gin.Context) {
	addr := c.Param("address")
	c.JSON(http.StatusOK, gin.H{"message": addr})
}

func GetBalance(c *gin.Context) {
	addr := c.Param("address")
	c.JSON(http.StatusOK, gin.H{"message": addr})
}