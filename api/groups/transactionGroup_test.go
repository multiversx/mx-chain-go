package groups_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataTx "github.com/multiversx/mx-chain-core-go/data/transaction"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/external"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewTransactionGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewTransactionGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type transactionResponseData struct {
	TxResp *groups.TxResponse `json:"transaction,omitempty"`
}

type transactionResponse struct {
	Data  transactionResponseData `json:"data"`
	Error string                  `json:"error"`
	Code  string                  `json:"code"`
}

type sendMultipleTxsResponseData struct {
	TxsSent   int      `json:"txsSent"`
	TxsHashes []string `json:"txsHashes"`
}

type sendMultipleTxsResponse struct {
	Data  sendMultipleTxsResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

type simulateTxResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  string      `json:"code"`
}

type sendSingleTxResponseData struct {
	TxHash string `json:"txHash"`
}

type sendSingleTxResponse struct {
	Data  sendSingleTxResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

type transactionCostResponseData struct {
	Cost uint64 `json:"txGasUnits"`
}

type transactionCostResponse struct {
	Data  transactionCostResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

type txsPoolResponseData struct {
	TxPool common.TransactionsPoolAPIResponse `json:"txPool"`
}

type txsPoolResponse struct {
	Data  txsPoolResponseData `json:"data"`
	Error string              `json:"error"`
	Code  string              `json:"code"`
}

type poolForSenderResponseData struct {
	TxPool common.TransactionsPoolForSenderApiResponse `json:"txPool"`
}

type poolForSenderResponse struct {
	Data  poolForSenderResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

type lastPoolNonceForSenderResponseData struct {
	Nonce uint64 `json:"nonce"`
}

type lastPoolNonceForSenderResponse struct {
	Data  lastPoolNonceForSenderResponseData `json:"data"`
	Error string                             `json:"error"`
	Code  string                             `json:"code"`
}

type txPoolNonceGapsForSenderResponseData struct {
	NonceGaps common.TransactionsPoolNonceGapsForSenderApiResponse `json:"nonceGaps"`
}

type txPoolNonceGapsForSenderResponse struct {
	Data  txPoolNonceGapsForSenderResponseData `json:"data"`
	Error string                               `json:"error"`
	Code  string                               `json:"code"`
}

func TestGetTransaction_WithCorrectHashShouldReturnTransaction(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := []byte("data")
	hash := "hash"
	guardian := "guardian"
	facade := mock.FacadeStub{
		GetTransactionHandler: func(hash string, withEvents bool) (i *dataTx.ApiTransactionResult, e error) {
			return &dataTx.ApiTransactionResult{
				Sender:       sender,
				Receiver:     receiver,
				Data:         txData,
				Value:        value,
				GuardianAddr: guardian,
			}, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/"+hash, nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := transactionResponse{}
	loadResponse(resp.Body, &response)

	txResp := response.Data.TxResp
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, txData, txResp.Data)
	assert.Equal(t, guardian, txResp.GuardianAddr)
}

func TestGetTransaction_WithUnknownHashShouldReturnNil(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := []byte("data")
	wrongHash := "wronghash"
	facade := mock.FacadeStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*dataTx.ApiTransactionResult, error) {
			if hash == wrongHash {
				return nil, errors.New("local error")
			}
			return &dataTx.ApiTransactionResult{
				Sender:   sender,
				Receiver: receiver,
				Data:     txData,
				Value:    value,
			}, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/"+wrongHash, nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := transactionResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Empty(t, txResp.Data)
}

func TestGetTransaction_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/eeee", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := transactionResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{}

	jsonBytes, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_WrongParametersShouldErrorOnValidation(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := "ishouldbeint"
	data := "data"

	facade := mock.FacadeStub{}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	jsonStr := fmt.Sprintf(`{"sender":"%s", "receiver":"%s", "value":%s, "data":"%s"}`,
		sender,
		receiver,
		value,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, txResp.Error, apiErrors.ErrValidation.Error())
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_ErrorWhenFacadeSendTransactionError(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "aabbccdd"
	errorString := "send transaction error"

	facade := mock.FacadeStub{
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return nil, nil, nil
		},
		SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (u uint64, err error) {
			return 0, errors.New(errorString)
		},
		ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	jsonStr := fmt.Sprintf(`{"sender":"%s", "receiver":"%s", "value":"%s", "signature":"%s", "data":"%s"}`,
		sender,
		receiver,
		value,
		signature,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, txResp.Error, errorString)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()
	nonce := uint64(1)
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "aabbccdd"
	hexTxHash := "deadbeef"

	facade := mock.FacadeStub{
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			txHash, _ := hex.DecodeString(hexTxHash)
			return nil, txHash, nil
		},
		SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (u uint64, err error) {
			return 1, nil
		},
		ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	jsonStr := fmt.Sprintf(
		`{"nonce": %d, "sender": "%s", "receiver": "%s", "value": "%s", "signature": "%s", "data": "%s"}`,
		nonce,
		sender,
		receiver,
		value,
		signature,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := sendSingleTxResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, response.Error)
	assert.Equal(t, hexTxHash, response.Data.TxHash)
}

func TestSendMultipleTransactions_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx0 := groups.SendTxRequest{}
	txs := []*groups.SendTxRequest{&tx0}

	jsonBytes, _ := json.Marshal(txs)
	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendMultipleTransactions_WrongPayloadShouldErrorOnValidation(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	jsonStr := `{"wrong": json}`

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, txResp.Error, apiErrors.ErrValidation.Error())
	assert.Empty(t, txResp.Data)
}

func TestSendMultipleTransactions_OkPayloadShouldWork(t *testing.T) {
	t.Parallel()

	createTxWasCalled := false
	sendBulkTxsWasCalled := false

	facade := mock.FacadeStub{
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			createTxWasCalled = true
			return &dataTx.Transaction{}, make([]byte, 0), nil
		},
		SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (u uint64, e error) {
			sendBulkTxsWasCalled = true
			return 0, nil
		},
		ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx0 := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	tx1 := tx0
	tx1.Sender = "sender2"
	txs := []*groups.SendTxRequest{&tx0, &tx1}

	jsonBytes, _ := json.Marshal(txs)

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txCostResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txCostResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, createTxWasCalled)
	assert.True(t, sendBulkTxsWasCalled)
}

func TestComputeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	expectedGasLimit := uint64(37)

	facade := mock.FacadeStub{
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return &dataTx.Transaction{}, nil, nil
		},
		ComputeTransactionGasLimitHandler: func(tx *dataTx.Transaction) (*dataTx.CostResponse, error) {
			return &dataTx.CostResponse{
				GasUnits:      expectedGasLimit,
				ReturnMessage: "",
			}, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx0 := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}

	jsonBytes, _ := json.Marshal(tx0)

	req, _ := http.NewRequest("POST", "/transaction/cost", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txCostResp := transactionCostResponse{}
	loadResponse(resp.Body, &txCostResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedGasLimit, txCostResp.Data.Cost)
}

func TestSimulateTransaction_BadRequestShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("POST", "/transaction/simulate", bytes.NewBuffer([]byte("invalid bytes")))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestSimulateTransaction_CreateErrorsShouldErr(t *testing.T) {
	t.Parallel()

	processTxWasCalled := false

	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResults, error) {
			processTxWasCalled = true
			return &txSimData.SimulationResults{
				Status:     "ok",
				FailReason: "no reason",
				ScResults:  nil,
				Receipts:   nil,
				Hash:       "hash",
			}, nil
		},
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return nil, nil, expectedErr
		},
		ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.False(t, processTxWasCalled)
	assert.Contains(t, simulateResponse.Error, expectedErr.Error())
}

func TestSimulateTransaction_ValidateErrorsShouldErr(t *testing.T) {
	t.Parallel()

	processTxWasCalled := false

	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResults, error) {
			processTxWasCalled = true
			return &txSimData.SimulationResults{
				Status:     "ok",
				FailReason: "no reason",
				ScResults:  nil,
				Receipts:   nil,
				Hash:       "hash",
			}, nil
		},
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return &dataTx.Transaction{}, []byte("hash"), nil
		},
		ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
			return expectedErr
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.False(t, processTxWasCalled)
	assert.Contains(t, simulateResponse.Error, expectedErr.Error())
}

func TestSimulateTransaction_CannotParseParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate?checkSignature=tttt", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, apiErrors.ErrValidation.Error(), simulateResponse.Error)
}

func TestSimulateTransaction_UseQueryParameterShouldWork(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
			assert.True(t, bypassSignature)
			return nil
		},
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return &dataTx.Transaction{}, []byte("hash"), nil
		},
		SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResults, error) {
			return &txSimData.SimulationResults{}, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate?bypassSignature=true", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSimulateTransaction_ProcessErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResults, error) {
			return nil, expectedErr
		},
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return &dataTx.Transaction{}, []byte("hash"), nil
		},
		ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, simulateResponse.Error, expectedErr.Error())
}

func TestSimulateTransaction(t *testing.T) {
	t.Parallel()

	processTxWasCalled := false

	facade := mock.FacadeStub{
		SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResults, error) {
			processTxWasCalled = true
			return &txSimData.SimulationResults{
				Status:     "ok",
				FailReason: "no reason",
				ScResults:  nil,
				Receipts:   nil,
				Hash:       "hash",
			}, nil
		},
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
			return &dataTx.Transaction{}, []byte("hash"), nil
		},
		ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
			return nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	tx := groups.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	jsonBytes, _ := json.Marshal(tx)

	req, _ := http.NewRequest("POST", "/transaction/simulate", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	simulateResponse := simulateTxResponse{}
	loadResponse(resp.Body, &simulateResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, processTxWasCalled)
	assert.Equal(t, string(shared.ReturnCodeSuccess), simulateResponse.Code)
}

func TestGetTransactionsPoolShouldError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
			return nil, expectedErr
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txsPoolResp := generalResponse{}
	loadResponse(resp.Body, &txsPoolResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(txsPoolResp.Error, expectedErr.Error()))
}

func TestGetTransactionsPoolShouldWork(t *testing.T) {
	t.Parallel()

	expectedTxPool := &common.TransactionsPoolAPIResponse{
		RegularTransactions: []common.Transaction{
			{
				TxFields: map[string]interface{}{
					"hash": "tx",
				},
			},
			{
				TxFields: map[string]interface{}{
					"hash": "tx2",
				},
			},
		},
	}
	facade := mock.FacadeStub{
		GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
			return expectedTxPool, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txsPoolResp := txsPoolResponse{}
	loadResponse(resp.Body, &txsPoolResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, txsPoolResp.Error)
	assert.Equal(t, *expectedTxPool, txsPoolResp.Data.TxPool)
}

func TestGetTransactionsPoolForSenderShouldError(t *testing.T) {
	t.Parallel()

	query := "?by-sender=sender"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
			return nil, expectedErr
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txsForSenderResp := poolForSenderResponse{}
	loadResponse(resp.Body, &txsForSenderResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(txsForSenderResp.Error, expectedErr.Error()))
}

func TestGetTransactionsPoolForSenderShouldWork(t *testing.T) {
	t.Parallel()

	expectedSender := "sender"
	providedFields := "sender,receiver"
	query := "?by-sender=" + expectedSender + "&fields=" + providedFields
	expectedResp := &common.TransactionsPoolForSenderApiResponse{
		Transactions: []common.Transaction{
			{
				TxFields: map[string]interface{}{
					"hash":     "txHash1",
					"sender":   expectedSender,
					"receiver": "receiver1",
				},
			},
			{
				TxFields: map[string]interface{}{
					"hash":     "txHash2",
					"sender":   expectedSender,
					"receiver": "receiver2",
				},
			},
		},
	}
	facade := mock.FacadeStub{
		GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
			return expectedResp, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txsForSenderResp := poolForSenderResponse{}
	loadResponse(resp.Body, &txsForSenderResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, txsForSenderResp.Error)
	assert.Equal(t, *expectedResp, txsForSenderResp.Data.TxPool)
}

func TestGetLastPoolNonceForSenderShouldError(t *testing.T) {
	t.Parallel()

	query := "?by-sender=sender&last-nonce=true"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
			return 0, expectedErr
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	lastPoolNonceResp := lastPoolNonceForSenderResponse{}
	loadResponse(resp.Body, &lastPoolNonceResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(lastPoolNonceResp.Error, expectedErr.Error()))
}

func TestGetLastPoolNonceForSenderShouldWork(t *testing.T) {
	t.Parallel()

	expectedSender := "sender"
	query := "?by-sender=" + expectedSender + "&last-nonce=true"
	expectedNonce := uint64(33)
	facade := mock.FacadeStub{
		GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
			return expectedNonce, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	lastPoolNonceResp := lastPoolNonceForSenderResponse{}
	loadResponse(resp.Body, &lastPoolNonceResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, lastPoolNonceResp.Error)
	assert.Equal(t, expectedNonce, lastPoolNonceResp.Data.Nonce)
}

func TestGetTransactionsPoolNonceGapsForSenderShouldError(t *testing.T) {
	t.Parallel()

	query := "?by-sender=sender&nonce-gaps=true"
	expectedErr := errors.New("expected error")
	facade := mock.FacadeStub{
		GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
			return nil, expectedErr
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	nonceGapsResp := txPoolNonceGapsForSenderResponse{}
	loadResponse(resp.Body, &nonceGapsResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(nonceGapsResp.Error, expectedErr.Error()))
}

func TestGetTransactionsPoolNonceGapsForSenderShouldWork(t *testing.T) {
	t.Parallel()

	expectedSender := "sender"
	query := "?by-sender=" + expectedSender + "&nonce-gaps=true"
	expectedNonceGaps := &common.TransactionsPoolNonceGapsForSenderApiResponse{
		Sender: expectedSender,
		Gaps: []common.NonceGapApiResponse{
			{
				From: 33,
				To:   60,
			},
		},
	}
	facade := mock.FacadeStub{
		GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
			return expectedNonceGaps, nil
		},
	}

	transactionGroup, err := groups.NewTransactionGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	nonceGapsResp := txPoolNonceGapsForSenderResponse{}
	loadResponse(resp.Body, &nonceGapsResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, nonceGapsResp.Error)
	assert.Equal(t, *expectedNonceGaps, nonceGapsResp.Data.NonceGaps)
}

func TestGetTransactionsPoolInvalidQueries(t *testing.T) {
	t.Parallel()

	t.Run("empty sender, requesting latest nonce", testTxPoolWithInvalidQuery("?last-nonce=true", apiErrors.ErrEmptySenderToGetLatestNonce))
	t.Run("empty sender, requesting nonce gaps", testTxPoolWithInvalidQuery("?nonce-gaps=true", apiErrors.ErrEmptySenderToGetNonceGaps))
	t.Run("fields + latest nonce", testTxPoolWithInvalidQuery("?fields=sender,receiver&last-nonce=true", apiErrors.ErrFetchingLatestNonceCannotIncludeFields))
	t.Run("fields + nonce gaps", testTxPoolWithInvalidQuery("?fields=sender,receiver&nonce-gaps=true", apiErrors.ErrFetchingNonceGapsCannotIncludeFields))
	t.Run("fields has spaces", testTxPoolWithInvalidQuery("?fields=sender ,receiver", apiErrors.ErrInvalidFields))
	t.Run("fields has numbers", testTxPoolWithInvalidQuery("?fields=sender1", apiErrors.ErrInvalidFields))
}

func testTxPoolWithInvalidQuery(query string, expectedErr error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		transactionGroup, err := groups.NewTransactionGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

		req, _ := http.NewRequest("GET", "/transaction/pool"+query, nil)

		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		txResp := &transactionResponse{}
		loadResponse(resp.Body, txResp)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrValidation.Error()))
		assert.True(t, strings.Contains(txResp.Error, expectedErr.Error()))
	}
}

func getTransactionRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"transaction": {
				Routes: []config.RouteConfig{
					{Name: "/send", Open: true},
					{Name: "/send-multiple", Open: true},
					{Name: "/cost", Open: true},
					{Name: "/pool", Open: true},
					{Name: "/:txhash", Open: true},
					{Name: "/:txhash/status", Open: true},
					{Name: "/simulate", Open: true},
				},
			},
		},
	}
}
