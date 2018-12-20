package transaction

import (
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error)
	GetTransaction(hash string) (*transaction.Transaction, error)
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type TxRequest struct {
	Sender   string   `form:"sender" json:"sender"`
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	Data     string   `form:"data" json:"data"`
	//SecretKey string `form:"sk" json:"sk" binding:"skValidator"`
}

//TxResponse represents the structure on which the response will be validated against
type TxResponse struct {
	Sender    string   `form:"sender" json:"sender"`
	Receiver  string   `form:"receiver" json:"receiver"`
	Value     *big.Int `form:"value" json:"value"`
	Data      string   `form:"data" json:"data"`
	Nonce     uint64   `form:"nonce" json:"nonce"`
	GasPrice  *big.Int `form:"gasPrice" json:"gasPrice"`
	GasLimit  *big.Int `form:"gasLimit" json:"gasLimit"`
	Signature string   `form:"signature" json:"signature"`
	Challenge string   `form:"challenge" json:"challenge"`
}

// Routes defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.POST("", GenerateTransaction)
	router.GET("/:txhash", GetTransaction)
}

// GenerateTransaction generates a new transaction given a sender, receiver, value and data
func GenerateTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	var gtx = TxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	tx, err := ef.GenerateTransaction(gtx.Sender, gtx.Receiver, *gtx.Value, gtx.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {

	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "TxHash is empty"})
	}

	tx, err := ef.GetTransaction(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction getting failed"})
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

func txResponseFromTransaction(tx *transaction.Transaction) TxResponse {
	response := TxResponse{}
	response.Nonce = tx.Nonce
	response.Sender = string(tx.SndAddr)
	response.Receiver = string(tx.RcvAddr)
	response.Data = string(tx.Data)
	response.Signature = string(tx.Signature)
	response.Challenge = string(tx.Challenge)

	//TODO: remove cast when tx.Value is bigInt
	response.Value = big.NewInt(int64(tx.Value))
	response.GasLimit = big.NewInt(int64(tx.GasLimit))
	response.GasPrice = big.NewInt(int64(tx.GasPrice))

	return response
}
