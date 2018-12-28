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

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/:address", GetAddress)
	router.GET("/:address/balance", GetBalance)
}

//GetAddress returns the information about the address passed as parameter
func GetAddress(c *gin.Context) {
	_, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	//TODO: add real implementation here
	addr := c.Param("address")

	c.JSON(http.StatusOK, gin.H{"message": addr})
}

//GetBalance returns the balance for the address parameter
func GetBalance(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}
	addr := c.Param("address")

	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Get balance error: Address was empty"})
		return
	}

	balance, err := ef.GetBalance(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Get balance error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}
