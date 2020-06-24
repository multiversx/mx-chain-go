package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/gin-gonic/gin"
)

// TxService interface defines methods that can be used from `elrondFacade` context variable
type TxService interface {
	CreateTransaction(nonce uint64, value string, receiver string, sender string, gasPrice uint64,
		gasLimit uint64, data string, signatureHex string) (*transaction.Transaction, []byte, error)
	ValidateTransaction(tx *transaction.Transaction) error
	SendBulkTransactions([]*transaction.Transaction) (uint64, error)
	GetTransaction(hash string) (*transaction.ApiTransactionResult, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error)
	EncodeAddressPubkey(pk []byte) (string, error)
	IsInterfaceNil() bool
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type TxRequest struct {
	Sender   string   `form:"sender" json:"sender"`
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	Data     string   `form:"data" json:"data"`
}

// MultipleTxRequest represents the structure on which user input for generating a bulk of transactions will validate against
type MultipleTxRequest struct {
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	TxCount  int      `form:"txCount" json:"txCount"`
}

// SendTxRequest represents the structure that maps and validates user input for publishing a new transaction
type SendTxRequest struct {
	Sender    string `form:"sender" json:"sender"`
	Receiver  string `form:"receiver" json:"receiver"`
	Value     string `form:"value" json:"value"`
	Data      string `form:"data" json:"data"`
	Nonce     uint64 `form:"nonce" json:"nonce"`
	GasPrice  uint64 `form:"gasPrice" json:"gasPrice"`
	GasLimit  uint64 `form:"gasLimit" json:"gasLimit"`
	Signature string `form:"signature" json:"signature"`
}

//TxResponse represents the structure on which the response will be validated against
type TxResponse struct {
	SendTxRequest
	ShardID     uint32 `json:"shardId"`
	Hash        string `json:"hash"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	Timestamp   uint64 `json:"timestamp"`
}

// Routes defines transaction related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodPost, "/send", SendTransaction)
	router.RegisterHandler(http.MethodPost, "/cost", ComputeTransactionGasLimit)
	router.RegisterHandler(http.MethodPost, "/send-multiple", SendMultipleTransactions)
	router.RegisterHandler(http.MethodGet, "/:txhash", GetTransaction)
}

// SendTransaction will receive a transaction from the client and propagate it for processing
func SendTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx = SendTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	tx, txHash, err := ef.CreateTransaction(
		gtx.Nonce,
		gtx.Value,
		gtx.Receiver,
		gtx.Sender,
		gtx.GasPrice,
		gtx.GasLimit,
		gtx.Data,
		gtx.Signature,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error())})
		return
	}

	err = ef.ValidateTransaction(tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error())})
		return
	}

	_, err = ef.SendBulkTransactions([]*transaction.Transaction{tx})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	txHexHash := hex.EncodeToString(txHash)
	c.JSON(http.StatusOK, gin.H{"txHash": txHexHash})
}

// SendMultipleTransactions will receive a number of transactions and will propagate them for processing
func SendMultipleTransactions(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx []SendTxRequest
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	var (
		txs    []*transaction.Transaction
		tx     *transaction.Transaction
		txHash []byte
	)

	txsHashes := make(map[int]string)
	for idx, receivedTx := range gtx {
		tx, txHash, err = ef.CreateTransaction(
			receivedTx.Nonce,
			receivedTx.Value,
			receivedTx.Receiver,
			receivedTx.Sender,
			receivedTx.GasPrice,
			receivedTx.GasLimit,
			receivedTx.Data,
			receivedTx.Signature,
		)
		if err != nil {
			continue
		}

		err = ef.ValidateTransaction(tx)
		if err != nil {
			continue
		}

		txs = append(txs, tx)
		txsHashes[idx] = hex.EncodeToString(txHash)
	}

	numOfSentTxs, err := ef.SendBulkTransactions(txs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"txsSent":   numOfSentTxs,
			"txsHashes": txsHashes,
		},
	)
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyTxHash.Error())})
		return
	}

	tx, err := ef.GetTransaction(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrTxNotFound.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": tx})
}

// ComputeTransactionGasLimit returns how many gas units a transaction wil consume
func ComputeTransactionGasLimit(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}
	var gtx SendTxRequest
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	tx, _, err := ef.CreateTransaction(
		gtx.Nonce,
		gtx.Value,
		gtx.Receiver,
		gtx.Sender,
		gtx.GasPrice,
		gtx.GasLimit,
		gtx.Data,
		gtx.Signature,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	cost, err := ef.ComputeTransactionGasLimit(tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"txGasUnits": cost})
}
