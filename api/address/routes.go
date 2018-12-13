package address

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	GetBalance(address string) (*big.Int, error)
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
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}
	addr := c.Param("address")

	balance, err := ef.GetBalance(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Get balance error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": balance})
}
