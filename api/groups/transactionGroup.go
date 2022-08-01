package groups

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/shared/logging"
	"github.com/ElrondNetwork/elrond-go/common"
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
	getTransactionsPool              = "/pool"

	queryParamWithResults    = "withResults"
	queryParamCheckSignature = "checkSignature"
	queryParamSender         = "by-sender"
	queryParamFields         = "fields"
	queryParamLastNonce      = "last-nonce"
	queryParamNonceGaps      = "nonce-gaps"
)

// transactionFacadeHandler defines the methods to be implemented by a facade for transaction requests
type transactionFacadeHandler interface {
	CreateTransaction(nonce uint64, value string, receiver string, receiverUsername []byte, sender string, senderUsername []byte, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32, options uint32) (*transaction.Transaction, []byte, error)
	ValidateTransaction(tx *transaction.Transaction) error
	ValidateTransactionForSimulation(tx *transaction.Transaction, checkSignature bool) error
	SendBulkTransactions([]*transaction.Transaction) (uint64, error)
	SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResults, error)
	GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error)
	GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSender(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error)
	EncodeAddressPubkey(pk []byte) (string, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
	IsInterfaceNil() bool
}

type transactionGroup struct {
	*baseGroup
	facade    transactionFacadeHandler
	mutFacade sync.RWMutex
}

// NewTransactionGroup returns a new instance of transactionGroup
func NewTransactionGroup(facade transactionFacadeHandler) (*transactionGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for transaction group", errors.ErrNilFacadeHandler)
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
					Position:   shared.Before,
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
					Position:   shared.Before,
				},
			},
		},
		{
			Path:    costPath,
			Method:  http.MethodPost,
			Handler: tg.computeTransactionGasLimit,
		},
		{
			Path:    getTransactionsPool,
			Method:  http.MethodGet,
			Handler: tg.getTransactionsPool,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(getTransactionPath, facade),
					Position:   shared.Before,
				},
			},
		},
		{
			Path:    sendMultiplePath,
			Method:  http.MethodPost,
			Handler: tg.sendMultipleTransactions,
			AdditionalMiddlewares: []shared.AdditionalMiddleware{
				{
					Middleware: middleware.CreateEndpointThrottlerFromFacade(sendMultipleTransactionsEndpoint, facade),
					Position:   shared.Before,
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
					Position:   shared.Before,
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

	start := time.Now()
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
	logging.LogAPIActionDurationIfNeeded(start, "API call: CreateTransaction")
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

	start = time.Now()
	err = tg.getFacade().ValidateTransactionForSimulation(tx, checkSignature)
	logging.LogAPIActionDurationIfNeeded(start, "API call: ValidateTransactionForSimulation")
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

	start = time.Now()
	executionResults, err := tg.getFacade().SimulateTransactionExecution(tx)
	logging.LogAPIActionDurationIfNeeded(start, "API call: SimulateTransactionExecution")
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

	start := time.Now()
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
	logging.LogAPIActionDurationIfNeeded(start, "API call: CreateTransaction")
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

	start = time.Now()
	err = tg.getFacade().ValidateTransaction(tx)
	logging.LogAPIActionDurationIfNeeded(start, "API call: ValidateTransaction")
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

	start = time.Now()
	_, err = tg.getFacade().SendBulkTransactions([]*transaction.Transaction{tx})
	logging.LogAPIActionDurationIfNeeded(start, "API call: SendBulkTransactions")
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

	var start time.Time
	txsHashes := make(map[int]string)
	for idx, receivedTx := range gtx {
		start = time.Now()
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
		logging.LogAPIActionDurationIfNeeded(start, "API call: CreateTransaction")
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

	start = time.Now()
	numOfSentTxs, err := tg.getFacade().SendBulkTransactions(txs)
	logging.LogAPIActionDurationIfNeeded(start, "API call: SendBulkTransactions")
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

	start := time.Now()
	tx, err := tg.getFacade().GetTransaction(txhash, withResults)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetTransaction")
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

	start := time.Now()
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
	logging.LogAPIActionDurationIfNeeded(start, "API call: CreateTransaction")
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

	start = time.Now()
	cost, err := tg.getFacade().ComputeTransactionGasLimit(tx)
	logging.LogAPIActionDurationIfNeeded(start, "API call: ComputeTransactionGasLimit")
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

// getTransactionsPool returns the transactions details in the pool
func (tg *transactionGroup) getTransactionsPool(c *gin.Context) {
	// extract and validate query parameters
	sender, fields, lastNonce, nonceGaps, err := tg.extractQueryParameters(c)
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

	err = validateQuery(sender, fields, lastNonce, nonceGaps)
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

	// if no sender was provided, the fields for all transactions from pool should be returned in response
	if sender == "" {
		tg.getTxPool(fields, c)
		return
	}

	if lastNonce {
		tg.getLastPoolNonceForSender(sender, c)
		return
	}

	if nonceGaps {
		tg.getTransactionsPoolNonceGapsForSender(sender, c)
		return
	}

	tg.getTxPoolForSender(sender, fields, c)
}

func (tg *transactionGroup) extractQueryParameters(c *gin.Context) (string, string, bool, bool, error) {
	senderAddress := getQueryParameterSender(c)
	fields := getQueryParameterFields(c)
	lastNonce, err := getQueryParameterLastNonce(c)
	if err != nil {
		return "", "", false, false, err
	}

	nonceGaps, err := getQueryParameterNonceGaps(c)
	if err != nil {
		return "", "", false, false, err
	}

	return senderAddress, fields, lastNonce, nonceGaps, nil
}

// getTxPool returns the fields for all txs in pool
func (tg *transactionGroup) getTxPool(fields string, c *gin.Context) {
	start := time.Now()
	txPool, err := tg.getFacade().GetTransactionsPool(fields)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetTransactionsPool")
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
			Data:  gin.H{"txPool": txPool},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getTxPoolForSender returns the fields for all txs in pool for the sender
func (tg *transactionGroup) getTxPoolForSender(sender, fields string, c *gin.Context) {
	start := time.Now()
	txPool, err := tg.getFacade().GetTransactionsPoolForSender(sender, fields)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetTransactionsPoolForSender")
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
			Data:  gin.H{"txPool": txPool},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getLastPoolNonceForSender returns the last nonce in pool for sender
func (tg *transactionGroup) getLastPoolNonceForSender(sender string, c *gin.Context) {
	start := time.Now()
	nonce, err := tg.getFacade().GetLastPoolNonceForSender(sender)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetLastPoolNonceForSender")
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
			Data:  gin.H{"nonce": nonce},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// getTransactionsPoolNonceGapsForSender returns the nonce gaps in pool for sender
func (tg *transactionGroup) getTransactionsPoolNonceGapsForSender(sender string, c *gin.Context) {
	start := time.Now()
	gaps, err := tg.getFacade().GetTransactionsPoolNonceGapsForSender(sender)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetTransactionsPoolNonceGapsForSender")
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
			Data:  gin.H{"nonceGaps": gaps},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func validateQuery(sender, fields string, lastNonce, nonceGaps bool) error {
	if fields != "" && lastNonce {
		return errors.ErrFetchingLatestNonceCannotIncludeFields
	}

	if fields != "" && nonceGaps {
		return errors.ErrFetchingNonceGapsCannotIncludeFields
	}

	if sender == "" && lastNonce {
		return errors.ErrEmptySenderToGetLatestNonce
	}

	if sender == "" && nonceGaps {
		return errors.ErrEmptySenderToGetNonceGaps
	}

	if fields != "" {
		return validateFields(fields)
	}

	return nil
}

func validateFields(fields string) error {
	for _, c := range fields {
		if c == ',' {
			continue
		}

		isLowerLetter := c >= 'a' && c <= 'z'
		isUpperLetter := c >= 'A' && c <= 'Z'
		if !isLowerLetter && !isUpperLetter {
			return errors.ErrInvalidFields
		}
	}

	return nil
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

func getQueryParameterSender(c *gin.Context) string {
	senderAddress := c.Request.URL.Query().Get(queryParamSender)
	return senderAddress
}

func getQueryParameterFields(c *gin.Context) string {
	fieldsStr := c.Request.URL.Query().Get(queryParamFields)
	return fieldsStr
}

func getQueryParameterLastNonce(c *gin.Context) (bool, error) {
	lastNonceStr := c.Request.URL.Query().Get(queryParamLastNonce)
	if lastNonceStr == "" {
		return false, nil
	}

	return strconv.ParseBool(lastNonceStr)
}

func getQueryParameterNonceGaps(c *gin.Context) (bool, error) {
	nonceGapsStr := c.Request.URL.Query().Get(queryParamNonceGaps)
	if nonceGapsStr == "" {
		return false, nil
	}

	return strconv.ParseBool(nonceGapsStr)
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
	castFacade, ok := newFacade.(transactionFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	tg.mutFacade.Lock()
	tg.facade = castFacade
	tg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tg *transactionGroup) IsInterfaceNil() bool {
	return tg == nil
}
