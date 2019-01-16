package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/gin-gonic/gin"
)

// TxService interface defines methods that can be used from `elrondFacade` context variable
type TxService interface {
	GenerateTransaction(sender string, receiver string, value big.Int, code string) (*transaction.Transaction, error)
	SendTransaction(nonce uint64, sender string, receiver string, value big.Int, code string, signature []byte) (*transaction.Transaction, error)
	GetTransaction(hash string) (*transaction.Transaction, error)
	GenerateAndSendBulkTransactions(string, big.Int, uint64) error
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type TxRequest struct {
	Sender   string   `form:"sender" json:"sender"`
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	Data     string   `form:"data" json:"data"`
	//SecretKey string `form:"sk" json:"sk" binding:"skValidator"`
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type MultipleTxRequest struct {
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	NoTxs    int      `form:"noTxs" json:"noTxs"`
}

// SendTxRequest represents the structure that maps and validates user input for publishing a new transaction
type SendTxRequest struct {
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

//TxResponse represents the structure on which the response will be validated against
type TxResponse struct {
	SendTxRequest
}

// Routes defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/generate", GenerateTransaction)
	router.POST("/generateAndSendMultiple", GenerateAndSendBulkTransactions)
	router.POST("/send", SendTransaction)
	router.GET("/:txhash", GetTransaction)
}

// GenerateTransaction generates a new transaction given a sender, receiver, value and data
func GenerateTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction generation failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// SendTransaction will receive a transaction from the client and propagate it for processing
func SendTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	var gtx = SendTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	signature, err := hex.DecodeString(gtx.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature, could not decode hex value: " + err.Error()})
		return
	}

	tx, err := ef.SendTransaction(gtx.Nonce, gtx.Sender, gtx.Receiver, *gtx.Value, gtx.Data, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction generation failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// GenerateAndSendBulkTransactions generates multipleTransactions
func GenerateAndSendBulkTransactions(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	var gtx = MultipleTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	err = ef.GenerateAndSendBulkTransactions(gtx.Receiver, *gtx.Value, uint64(gtx.NoTxs))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Multiple Transaction generation failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d", gtx.NoTxs)})
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {

	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TxHash is empty"})
		return
	}

	tx, err := ef.GetTransaction(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction getting failed"})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction was not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

func txResponseFromTransaction(tx *transaction.Transaction) TxResponse {
	response := TxResponse{}
	response.Nonce = tx.Nonce
	response.Sender = hex.EncodeToString(tx.SndAddr)
	response.Receiver = hex.EncodeToString(tx.RcvAddr)
	response.Data = string(tx.Data)
	response.Signature = hex.EncodeToString(tx.Signature)
	response.Challenge = string(tx.Challenge)
	response.Value = &tx.Value
	response.GasLimit = big.NewInt(int64(tx.GasLimit))
	response.GasPrice = big.NewInt(int64(tx.GasPrice))

	return response
}
