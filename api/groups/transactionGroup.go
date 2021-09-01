package groups

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/gin-gonic/gin"
)

const (
	sendTransactionEndpoint          = "/transaction/send"
	simulateTransactionEndpoint      = "/transaction/simulate"
	sendMultipleTransactionsEndpoint = "/transaction/send-multiple"
	getTransactionEndpoint           = "/transaction/:hash"
	sendTransactionPath              = "/send"
	simulateTransactionPath          = "/simulate"
	costPath                         = "/cost"
	sendMultiplePath                 = "/send-multiple"
	getTransactionPath               = "/:txhash"

	queryParamWithResults    = "withResults"
	queryParamCheckSignature = "checkSignature"
)

// transactionFacadeHandler interface defines methods that can be used by the gin webserver
type transactionFacadeHandler interface {
	CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransaction(tx *transaction.Transaction) error
	ValidateTransactionForSimulation(tx *transaction.Transaction, checkSignature bool) error
	SendBulkTransactions([]*transaction.Transaction) (uint64, error)
	SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	EncodeAddressPubkey(pk []byte) (string, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
	IsInterfaceNil() bool
}

type transactionGroup struct {
	facade    transactionFacadeHandler
	mutFacade sync.RWMutex
	*baseGroup
}

// NewTransactionGroup returns a new instance of transactionGroup
func NewTransactionGroup(facadeHandler interface{}) (*transactionGroup, error) {
	if facadeHandler == nil {
		return nil, errors.ErrNilFacadeHandler
	}

	facade, ok := facadeHandler.(transactionFacadeHandler)
	if !ok {
		return nil, fmt.Errorf("%w for transaction group", errors.ErrFacadeWrongTypeAssertion)
	}

	tg := &transactionGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    sendTransactionPath,
			Method:  http.MethodPost,
			Handler: tg.sendTransaction,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(sendTransactionEndpoint, facade),
					Before:     true,
				},
			},
		},
		{
			Path:    simulateTransactionPath,
			Method:  http.MethodPost,
			Handler: tg.simulateTransaction,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(simulateTransactionEndpoint, facade),
					Before:     true,
				},
			},
		},
		{
			Path:    costPath,
			Method:  http.MethodPost,
			Handler: tg.computeTransactionGasLimit,
		},
		{
			Path:    sendMultiplePath,
			Method:  http.MethodPost,
			Handler: tg.sendMultipleTransactions,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(sendMultipleTransactionsEndpoint, facade),
					Before:     true,
				},
			},
		},
		{
			Path:    getTransactionPath,
			Method:  http.MethodGet,
			Handler: tg.getTransaction,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(getTransactionEndpoint, facade),
					Before:     true,
				},
			},
		},
	}
	tg.endpoints = endpoints

	return tg, nil
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
	Sender           string `form:"sender" json:"sender"`
	Receiver         string `form:"receiver" json:"receiver"`
	SenderUsername   []byte `json:"senderUsername,omitempty"`
	ReceiverUsername []byte `json:"receiverUsername,omitempty"`
	Value            string `form:"value" json:"value"`
	Data             []byte `form:"data" json:"data"`
	Nonce            uint64 `form:"nonce" json:"nonce"`
	GasPrice         uint64 `form:"gasPrice" json:"gasPrice"`
	GasLimit         uint64 `form:"gasLimit" json:"gasLimit"`
	Signature        string `form:"signature" json:"signature"`
	ChainID          string `form:"chainID" json:"chainID"`
	Version          uint32 `form:"version" json:"version"`
	Options          uint32 `json:"options,omitempty"`
}

// TxResponse represents the structure on which the response will be validated against
type TxResponse struct {
	SendTxRequest
	ShardID     uint32 `json:"shardId"`
	Hash        string `json:"hash"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	Timestamp   uint64 `json:"timestamp"`
}

// simulateTransaction will receive a transaction from the client and will simulate it's execution and return the results
func (tg *transactionGroup) simulateTransaction(c *gin.Context) {
	var gtx = SendTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	checkSignature, err := getQueryParameterCheckSignature(c)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrValidation.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tx, txHash, err := tg.getFacade().CreateTransaction(
		gtx.Nonce,
		gtx.Value,
		gtx.Receiver,
		gtx.ReceiverUsername,
		gtx.Sender,
		gtx.SenderUsername,
		gtx.GasPrice,
		gtx.GasLimit,
		gtx.Data,
		gtx.Signature,
		gtx.ChainID,
		gtx.Version,
		gtx.Options,
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	err = tg.getFacade().ValidateTransactionForSimulation(tx, checkSignature)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	executionResults, err := tg.getFacade().SimulateTransactionExecution(tx)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	executionResults.Hash = hex.EncodeToString(txHash)
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"result": executionResults},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// sendTransaction will receive a transaction from the client and propagate it for processing
func (tg *transactionGroup) sendTransaction(c *gin.Context) {
	var gtx = SendTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tx, txHash, err := tg.getFacade().CreateTransaction(
		gtx.Nonce,
		gtx.Value,
		gtx.Receiver,
		gtx.ReceiverUsername,
		gtx.Sender,
		gtx.SenderUsername,
		gtx.GasPrice,
		gtx.GasLimit,
		gtx.Data,
		gtx.Signature,
		gtx.ChainID,
		gtx.Version,
		gtx.Options,
	)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	err = tg.getFacade().ValidateTransaction(tx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	_, err = tg.getFacade().SendBulkTransactions([]*transaction.Transaction{tx})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	txHexHash := hex.EncodeToString(txHash)
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"txHash": txHexHash},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// sendMultipleTransactions will receive a number of transactions and will propagate them for processing
func (tg *transactionGroup) sendMultipleTransactions(c *gin.Context) {
	var gtx []SendTxRequest
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	var (
		txs    []*transaction.Transaction
		tx     *transaction.Transaction
		txHash []byte
	)

	txsHashes := make(map[int]string)
	for idx, receivedTx := range gtx {
		tx, txHash, err = tg.getFacade().CreateTransaction(
			receivedTx.Nonce,
			receivedTx.Value,
			receivedTx.Receiver,
			receivedTx.ReceiverUsername,
			receivedTx.Sender,
			receivedTx.SenderUsername,
			receivedTx.GasPrice,
			receivedTx.GasLimit,
			receivedTx.Data,
			receivedTx.Signature,
			receivedTx.ChainID,
			receivedTx.Version,
			receivedTx.Options,
		)
		if err != nil {
			continue
		}

		err = tg.getFacade().ValidateTransaction(tx)
		if err != nil {
			continue
		}

		txs = append(txs, tx)
		txsHashes[idx] = hex.EncodeToString(txHash)
	}

	numOfSentTxs, err := tg.getFacade().SendBulkTransactions(txs)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data: gin.H{
				"txsSent":   numOfSentTxs,
				"txsHashes": txsHashes,
			},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getTransaction returns transaction details for a given txhash
func (tg *transactionGroup) getTransaction(c *gin.Context) {
	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyTxHash.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	withResults, err := getQueryParamWithResults(c)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrValidation.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tx, err := tg.getFacade().GetTransaction(txhash, withResults)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetTransaction.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"transaction": tx},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// computeTransactionGasLimit returns how many gas units a transaction wil consume
func (tg *transactionGroup) computeTransactionGasLimit(c *gin.Context) {
	var gtx SendTxRequest
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tx, _, err := tg.getFacade().CreateTransaction(
		gtx.Nonce,
		gtx.Value,
		gtx.Receiver,
		gtx.ReceiverUsername,
		gtx.Sender,
		gtx.SenderUsername,
		gtx.GasPrice,
		gtx.GasLimit,
		gtx.Data,
		gtx.Signature,
		gtx.ChainID,
		gtx.Version,
		gtx.Options,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	cost, err := tg.getFacade().ComputeTransactionGasLimit(tx)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  cost,
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func getQueryParamWithResults(c *gin.Context) (bool, error) {
	withResultsStr := c.Request.URL.Query().Get(queryParamWithResults)
	if withResultsStr == "" {
		return false, nil
	}

	return strconv.ParseBool(withResultsStr)
}

func getQueryParameterCheckSignature(c *gin.Context) (bool, error) {
	bypassSignatureStr := c.Request.URL.Query().Get(queryParamCheckSignature)
	if bypassSignatureStr == "" {
		return true, nil
	}

	return strconv.ParseBool(bypassSignatureStr)
}

func (tg *transactionGroup) getFacade() transactionFacadeHandler {
	tg.mutFacade.RLock()
	defer tg.mutFacade.RUnlock()

	return tg.facade
}

// UpdateFacade will update the facade
func (tg *transactionGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castedFacade, ok := newFacade.(transactionFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	tg.mutFacade.Lock()
	tg.facade = castedFacade
	tg.mutFacade.Unlock()

	return nil
}
