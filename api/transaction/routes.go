package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	GenerateTransaction(sender string, receiver string, amount big.Int, code string) (string, error)
	GetTransaction(hash string) (transaction.Transaction, error)
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type TxRequest struct {
	Sender   string  `form:"sender" json:"sender"`
	Receiver string  `form:"receiver" json:"receiver"`
	Amount   big.Int `form:"amount" json:"amount"`
	Code     string  `form:"code" json:"code"`
	//SecretKey string `form:"sk" json:"sk" binding:"skValidator"`
}

type TxResponse struct {
	Sender    string  `form:"sender" json:"sender"`
	Receiver  string  `form:"receiver" json:"receiver"`
	Amount    big.Int `form:"amount" json:"amount"`
	Data      string  `form:"data" json:"data"`
	Nonce     uint64  `form:"nonce" json:"nonce"`
	GasPrice  big.Int `form:"gasPrice" json:"gasPrice"`
	GasLimit  big.Int `form:"gasLimit" json:"gasLimit"`
	Signature string  `form:"signature" json:"signature"`
	Challenge string  `form:"challenge" json:"challenge"`
}

// Routes function defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/", GenerateTransaction)
	router.GET("/:txhash", GetTransaction)
}

// GenerateTransaction generates a new transaction given an sk and additional data
func GenerateTransaction(c *gin.Context) {
	var gtx = TxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Validation error: " + err.Error()})
		return
	}
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}
	hash, err := ef.GenerateTransaction(gtx.Sender, gtx.Receiver, gtx.Amount, gtx.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Transaction generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "hash": hash})
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {
	txhash := c.Param("txhash")
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid app context"})
		return
	}

	tx, err := ef.GetTransaction(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Transaction getting failed"})
	}
	var response = TxResponseFromTransaction(tx)
	c.JSON(http.StatusOK, gin.H{"message": "ok", "transaction": response})
}

func TxResponseFromTransaction(tx transaction.Transaction) TxResponse {
	response := TxResponse{}
	response.Nonce = tx.Nonce
	response.Sender = string(tx.SndAddr)
	response.Receiver = string(tx.RcvAddr)
	response.Data = string(tx.Data)
	response.Signature = string(tx.Signature)
	response.Challenge = string(tx.Challenge)

	//TODO: remove cast when tx.Value is bigInt
	response.Amount = *big.NewInt(int64(tx.Value))
	response.GasLimit = *big.NewInt(int64(tx.GasLimit))
	response.GasPrice = *big.NewInt(int64(tx.GasPrice))

	return response
}
