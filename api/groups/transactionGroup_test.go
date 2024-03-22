package groups_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
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

var (
	sender      = "sender"
	receiver    = "receiver"
	value       = "10"
	txData      = []byte("data")
	hash        = "hash"
	guardian    = "guardian"
	signature   = "aabbccdd"
	expectedErr = errors.New("expected error")
	nonce       = uint64(1)
	hexTxHash   = "deadbeef"
	jsonTxStr   = fmt.Sprintf(`{"nonce": %d, "sender":"%s", "receiver":"%s", "value":"%s", "signature":"%s", "data":"%s"}`,
		nonce,
		sender,
		receiver,
		value,
		signature,
		txData,
	)
)

func TestTransactionsGroup_getTransaction(t *testing.T) {
	t.Parallel()

	t.Run("number of go routines exceeded", testExceededNumGoRoutines("/transaction/eeee", nil))
	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetTransactionHandler: func(hash string, withEvents bool) (*dataTx.ApiTransactionResult, error) {
				require.Fail(t, "should have not been called")
				return &dataTx.ApiTransactionResult{}, nil
			},
		}

		transactionGroup, err := groups.NewTransactionGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

		req, _ := http.NewRequest("GET", "/transaction/hash?withResults=not-bool", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		txResp := transactionResponse{}
		loadResponse(resp.Body, &txResp)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Empty(t, txResp.Data)
	})
	t.Run("facade returns error should error", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetTransactionHandler: func(hash string, withEvents bool) (*dataTx.ApiTransactionResult, error) {
				return nil, expectedErr
			},
		}

		transactionGroup, err := groups.NewTransactionGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

		req, _ := http.NewRequest("GET", "/transaction/hash", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		txResp := transactionResponse{}
		loadResponse(resp.Body, &txResp)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Empty(t, txResp.Data)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
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

		response := &transactionResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/"+hash,
			"GET",
			nil,
			response,
		)
		txResp := response.Data.TxResp
		assert.Equal(t, sender, txResp.Sender)
		assert.Equal(t, receiver, txResp.Receiver)
		assert.Equal(t, value, txResp.Value)
		assert.Equal(t, txData, txResp.Data)
		assert.Equal(t, guardian, txResp.GuardianAddr)
	})
}

func TestTransactionGroup_sendTransaction(t *testing.T) {
	t.Parallel()

	t.Run("number of go routines exceeded", testExceededNumGoRoutines("/transaction/send", &dataTx.FrontendTransaction{}))
	t.Run("invalid params should error", testTransactionGroupErrorScenario("/transaction/send", "POST", jsonTxStr, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("CreateTransaction error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, expectedErr
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				require.Fail(t, "should have not been called")
				return nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusBadRequest,
			expectedErr,
		)
	})
	t.Run("ValidateTransaction error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				return expectedErr
			},
			SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (u uint64, err error) {
				require.Fail(t, "should have not been called")
				return 0, nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusBadRequest,
			expectedErr,
		)
	})
	t.Run("SendBulkTransactions error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (u uint64, err error) {
				return 0, expectedErr
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				return nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
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

		response := &sendSingleTxResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/send",
			"POST",
			bytes.NewBuffer([]byte(jsonTxStr)),
			response,
		)
		assert.Empty(t, response.Error)
		assert.Equal(t, hexTxHash, response.Data.TxHash)
	})
}

func TestTransactionGroup_sendMultipleTransactions(t *testing.T) {
	t.Parallel()

	t.Run("number of go routines exceeded", testExceededNumGoRoutines("/transaction/send-multiple", &dataTx.FrontendTransaction{}))
	t.Run("invalid params should error", testTransactionGroupErrorScenario("/transaction/send-multiple", "POST", jsonTxStr, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("CreateTransaction error should continue, error on SendBulkTransactions", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, expectedErr
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				require.Fail(t, "should not have been called")
				return nil
			},
			SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (uint64, error) {
				require.Zero(t, len(txs))
				return 0, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send-multiple",
			"POST",
			[]*dataTx.FrontendTransaction{{}},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("ValidateTransaction error should continue, error on SendBulkTransactions", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				return expectedErr
			},
			SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (uint64, error) {
				require.Zero(t, len(txs))
				return 0, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send-multiple",
			"POST",
			[]*dataTx.FrontendTransaction{{}},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("SendBulkTransactions error error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ValidateTransactionHandler: func(tx *dataTx.Transaction) error {
				return nil
			},
			SendBulkTransactionsHandler: func(txs []*dataTx.Transaction) (uint64, error) {
				require.Equal(t, 1, len(txs))
				return 0, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/send-multiple",
			"POST",
			[]*dataTx.FrontendTransaction{{}},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		createTxWasCalled := false
		sendBulkTxsWasCalled := false

		facade := &mock.FacadeStub{
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

		tx0 := dataTx.FrontendTransaction{
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
		txs := []*dataTx.FrontendTransaction{&tx0, &tx1}

		jsonBytes, _ := json.Marshal(txs)

		response := &sendMultipleTxsResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/send-multiple",
			"POST",
			bytes.NewBuffer(jsonBytes),
			response,
		)
		assert.True(t, createTxWasCalled)
		assert.True(t, sendBulkTxsWasCalled)
	})
}

func TestTransactionGroup_computeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", testTransactionGroupErrorScenario("/transaction/cost", "POST", jsonTxStr, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("CreateTransaction error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, expectedErr
			},
			ComputeTransactionGasLimitHandler: func(tx *dataTx.Transaction) (*dataTx.CostResponse, error) {
				require.Fail(t, "should not have been called")
				return nil, nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/cost",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("ComputeTransactionGasLimit error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ComputeTransactionGasLimitHandler: func(tx *dataTx.Transaction) (*dataTx.CostResponse, error) {
				return nil, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/cost",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedGasLimit := uint64(37)

		facade := &mock.FacadeStub{
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

		tx0 := dataTx.FrontendTransaction{
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

		response := &transactionCostResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/cost",
			"POST",
			bytes.NewBuffer(jsonBytes),
			response,
		)
		assert.Equal(t, expectedGasLimit, response.Data.Cost)
	})
}

func TestTransactionGroup_simulateTransaction(t *testing.T) {
	t.Parallel()

	t.Run("number of go routines exceeded", testExceededNumGoRoutines("/transaction/simulate", &dataTx.FrontendTransaction{}))
	t.Run("invalid param transaction should error", testTransactionGroupErrorScenario("/transaction/simulate", "POST", jsonTxStr, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("invalid param checkSignature should error", testTransactionGroupErrorScenario("/transaction/simulate?checkSignature=not-bool", "POST", &dataTx.FrontendTransaction{}, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("CreateTransaction error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, expectedErr
			},
			ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
				require.Fail(t, "should have not been called")
				return nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/simulate",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusBadRequest,
			expectedErr,
		)
	})
	t.Run("ValidateTransactionForSimulation error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
				return expectedErr
			},
			SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
				require.Fail(t, "should have not been called")
				return nil, nil
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/simulate",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusBadRequest,
			expectedErr,
		)
	})
	t.Run("SimulateTransactionExecution error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return nil, nil, nil
			},
			ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
				return nil
			},
			SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
				return nil, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/simulate",
			"POST",
			&dataTx.FrontendTransaction{},
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processTxWasCalled := false

		facade := &mock.FacadeStub{
			SimulateTransactionExecutionHandler: func(tx *dataTx.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
				processTxWasCalled = true
				return &txSimData.SimulationResultsWithVMOutput{
					SimulationResults: dataTx.SimulationResults{
						Status:     "ok",
						FailReason: "no reason",
						ScResults:  nil,
						Receipts:   nil,
						Hash:       "hash",
					},
				}, nil
			},
			CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*dataTx.Transaction, []byte, error) {
				return &dataTx.Transaction{}, []byte("hash"), nil
			},
			ValidateTransactionForSimulationHandler: func(tx *dataTx.Transaction, bypassSignature bool) error {
				return nil
			},
		}

		tx := dataTx.FrontendTransaction{
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

		response := &simulateTxResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/simulate",
			"POST",
			bytes.NewBuffer(jsonBytes),
			response,
		)
		assert.True(t, processTxWasCalled)
		assert.Equal(t, string(shared.ReturnCodeSuccess), response.Code)
	})
}

func TestTransactionGroup_getTransactionsPool(t *testing.T) {
	t.Parallel()

	t.Run("number of go routines exceeded", testExceededNumGoRoutines("/transaction/pool", nil))
	t.Run("invalid last-nonce param should error", testTransactionGroupErrorScenario("/transaction/pool?last-nonce=not-bool", "GET", nil, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("invalid nonce-gaps param should error", testTransactionGroupErrorScenario("/transaction/pool?nonce-gaps=not-bool", "GET", nil, http.StatusBadRequest, apiErrors.ErrValidation))
	t.Run("empty sender, requesting latest nonce", testTxPoolWithInvalidQuery("?last-nonce=true", apiErrors.ErrEmptySenderToGetLatestNonce))
	t.Run("empty sender, requesting nonce gaps", testTxPoolWithInvalidQuery("?nonce-gaps=true", apiErrors.ErrEmptySenderToGetNonceGaps))
	t.Run("fields + latest nonce", testTxPoolWithInvalidQuery("?fields=sender,receiver&last-nonce=true", apiErrors.ErrFetchingLatestNonceCannotIncludeFields))
	t.Run("fields + nonce gaps", testTxPoolWithInvalidQuery("?fields=sender,receiver&nonce-gaps=true", apiErrors.ErrFetchingNonceGapsCannotIncludeFields))
	t.Run("fields has spaces", testTxPoolWithInvalidQuery("?fields=sender ,receiver", apiErrors.ErrInvalidFields))
	t.Run("fields has numbers", testTxPoolWithInvalidQuery("?fields=sender1", apiErrors.ErrInvalidFields))
	t.Run("fields + wild card", testTxPoolWithInvalidQuery("?fields=sender,receiver,*", apiErrors.ErrInvalidFields))
	t.Run("GetTransactionsPool error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/pool",
			"GET",
			nil,
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("GetTransactionsPoolForSender error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/pool?by-sender=sender",
			"GET",
			nil,
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("GetLastPoolNonceForSender error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return 0, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/pool?by-sender=sender&last-nonce=true",
			"GET",
			nil,
			http.StatusInternalServerError,
			expectedErr,
		)
	})
	t.Run("GetTransactionsPoolNonceGapsForSender error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}
		testTransactionsGroup(
			t,
			facade,
			"/transaction/pool?by-sender=sender&nonce-gaps=true",
			"GET",
			nil,
			http.StatusInternalServerError,
			expectedErr,
		)
	})

	t.Run("should work", func(t *testing.T) {
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
		facade := &mock.FacadeStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return expectedTxPool, nil
			},
		}

		response := &txsPoolResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/pool",
			"GET",
			nil,
			response,
		)
		assert.Empty(t, response.Error)
		assert.Equal(t, *expectedTxPool, response.Data.TxPool)
	})
	t.Run("should work for sender", func(t *testing.T) {
		t.Parallel()

		expectedSender := "sender"
		query := "?by-sender=" + expectedSender + "&fields=*"
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
		facade := &mock.FacadeStub{
			GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
				return expectedResp, nil
			},
		}

		response := &poolForSenderResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/pool"+query,
			"GET",
			nil,
			response,
		)
		assert.Empty(t, response.Error)
		assert.Equal(t, *expectedResp, response.Data.TxPool)
	})
	t.Run("should work for last pool nonce", func(t *testing.T) {
		t.Parallel()

		expectedSender := "sender"
		query := "?by-sender=" + expectedSender + "&last-nonce=true"
		expectedNonce := uint64(33)
		facade := &mock.FacadeStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return expectedNonce, nil
			},
		}

		response := &lastPoolNonceForSenderResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/pool"+query,
			"GET",
			nil,
			response,
		)
		assert.Empty(t, response.Error)
		assert.Equal(t, expectedNonce, response.Data.Nonce)
	})
	t.Run("should work for nonce gaps", func(t *testing.T) {
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
		facade := &mock.FacadeStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return expectedNonceGaps, nil
			},
		}

		response := &txPoolNonceGapsForSenderResponse{}
		loadTransactionGroupResponse(
			t,
			facade,
			"/transaction/pool"+query,
			"GET",
			nil,
			response,
		)
		assert.Empty(t, response.Error)
		assert.Equal(t, *expectedNonceGaps, response.Data.NonceGaps)
	})
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

func TestTransactionsGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		transactionGroup, err := groups.NewTransactionGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = transactionGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		transactionGroup, err := groups.NewTransactionGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = transactionGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTxPool := &common.TransactionsPoolAPIResponse{
			RegularTransactions: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "tx",
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

		newFacade := mock.FacadeStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}

		err = transactionGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/transaction/pool", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		loadResponse(resp.Body, &txsPoolResp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(txsPoolResp.Error, expectedErr.Error()))
	})
}

func TestTransactionsGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	transactionGroup, _ := groups.NewTransactionGroup(nil)
	require.True(t, transactionGroup.IsInterfaceNil())

	transactionGroup, _ = groups.NewTransactionGroup(&mock.FacadeStub{})
	require.False(t, transactionGroup.IsInterfaceNil())
}

func loadTransactionGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	transactionGroup, err := groups.NewTransactionGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func testTransactionGroupErrorScenario(
	url string,
	method string,
	body interface{},
	expectedCode int,
	expectedError error,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testTransactionsGroup(
			t,
			&mock.FacadeStub{},
			url,
			method,
			body,
			expectedCode,
			expectedError)
	}
}

func testExceededNumGoRoutines(url string, body interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		facade := &mock.FacadeStub{
			GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
				return &mock.ThrottlerStub{
					CanProcessCalled: func() bool { return false },
				}, true
			},
		}

		testTransactionsGroup(
			t,
			facade,
			url,
			"GET",
			body,
			http.StatusTooManyRequests,
			apiErrors.ErrTooManyRequests,
		)
	}
}

func testTransactionsGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body interface{},
	expectedRespCode int,
	expectedRespError error,
) {
	transactionGroup, err := groups.NewTransactionGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(transactionGroup, "transaction", getTransactionRoutesConfig())

	jsonBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, expectedRespCode, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, expectedRespError.Error()))
	assert.Empty(t, txResp.Data)
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
