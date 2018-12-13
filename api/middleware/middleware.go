package middleware

import (
	"github.com/gin-gonic/gin"
)

// ElrondHandler interface defines methods that can be used from `elrondFacade` context variable
type ElrondHandler interface {
	//IsNodeRunning() bool
	//StartNode() error
	//StopNode() error
	//GetBalance(address string) (*big.Int, error)
	//GenerateTransaction(sender string, receiver string, amount big.Int, code string) (string, error)
	//GetTransaction(hash string) (*transaction.Transaction, error)
}

// WithElrondFacade middleware will set up an ElrondFacade object in the gin context
func WithElrondFacade(elrondFacade ElrondHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondFacade", elrondFacade)
		c.Next()
	}
}
