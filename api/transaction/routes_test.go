package transaction_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	errors2 "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	tr "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type TransactionResponse struct {
	GeneralResponse
	TxResp *transaction.TxResponse `json:"transaction,omitempty"`
}

type TransactionHashResponse struct {
	GeneralResponse
	TxHash string `json:"txHash,omitempty"`
}

type transactionCostResponseData struct {
	Cost uint64 `json:"txGasUnits"`
}

type TransactionCostResponse struct {
	Data  transactionCostResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetTransaction_WithCorrectHashShouldReturnTransaction(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := "data"
	hash := "hash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.ApiTransactionResult, e error) {
			return &tr.ApiTransactionResult{
				Sender:   sender,
				Receiver: receiver,
				Data:     txData,
				Value:    value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+hash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	transactionResponse := TransactionResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &transactionResponse)

	txResp := transactionResponse.TxResp
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, txData, txResp.Data)
}

func TestGetTransaction_WithUnknownHashShouldReturnNil(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := "data"
	wrongHash := "wronghash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (*tr.ApiTransactionResult, error) {
			if hash == wrongHash {
				return nil, errors.New("local error")
			}
			return &tr.ApiTransactionResult{
				Sender:   sender,
				Receiver: receiver,
				Data:     txData,
				Value:    value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+wrongHash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Nil(t, transactionResponse.TxResp)
}

func TestGetTransaction_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/transaction/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, transactionResponse.Error, errors2.ErrInvalidAppContext.Error())
}

func TestSendTransaction_ErrorWithWrongFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("POST", "/transaction/send", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, transactionResponse.Error, errors2.ErrInvalidAppContext.Error())
}

func TestSendTransaction_WrongParametersShouldErrorOnValidation(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := "ishouldbeint"
	data := "data"

	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(`{"sender":"%s", "receiver":"%s", "value":%s, "data":"%s"}`,
		sender,
		receiver,
		value,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrValidation.Error())
	assert.Empty(t, transactionResponse.TxResp)
}

func TestSendTransaction_ErrorWhenFacadeSendTransactionError(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "aabbccdd"
	errorString := "send transaction error"

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ string, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			return nil, nil, nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, err error) {
			return 0, errors.New(errorString)
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

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

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, errorString)
	assert.Empty(t, transactionResponse.TxResp)
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

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ string, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			txHash, _ := hex.DecodeString(hexTxHash)
			return nil, txHash, nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, err error) {
			return 1, nil
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

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

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	txHashResponse := TransactionHashResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &txHashResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, txHashResponse.Error)
	assert.Equal(t, hexTxHash, txHashResponse.TxHash)
}

func TestSendMultipleTransactions_ErrorWithWrongFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("POST", "/transaction/send-multiple", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, transactionResponse.Error, errors2.ErrInvalidAppContext.Error())
}

func TestSendMultipleTransactions_WrongPayloadShouldErrorOnValidation(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	jsonStr := `{"wrong": json}`

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrValidation.Error())
	assert.Empty(t, transactionResponse.TxResp)
}

func TestSendMultipleTransactions_OkPayloadShouldWork(t *testing.T) {
	t.Parallel()

	createTxWasCalled := false
	sendBulkTxsWasCalled := false

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ string, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			createTxWasCalled = true
			return &tr.Transaction{}, make([]byte, 0), nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, e error) {
			sendBulkTxsWasCalled = true
			return 0, nil
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

	tx0 := transaction.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      "",
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	tx1 := tx0
	tx1.Sender = "sender2"
	txs := []*transaction.SendTxRequest{&tx0, &tx1}

	jsonBytes, _ := json.Marshal(txs)

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, createTxWasCalled)
	assert.True(t, sendBulkTxsWasCalled)
}

func TestComputeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	expectedGasLimit := uint64(37)

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ string, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			return &tr.Transaction{}, nil, nil
		},
		ComputeTransactionGasLimitHandler: func(tx *tr.Transaction) (uint64, error) {
			return expectedGasLimit, nil
		},
	}
	ws := startNodeServer(&facade)

	tx0 := transaction.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      "",
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}

	jsonBytes, _ := json.Marshal(tx0)

	req, _ := http.NewRequest("POST", "/transaction/cost", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionCostResponse := TransactionCostResponse{}
	loadResponse(resp.Body, &transactionCostResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedGasLimit, transactionCostResponse.Data.Cost)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler transaction.TxService) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ginTransactionRoute := ws.Group("/transaction")
	if handler != nil {
		ginTransactionRoute.Use(middleware.WithElrondFacade(handler))
	}
	transactionRoute, _ := wrapper.NewRouterWrapper("transaction", ginTransactionRoute, getRoutesConfig())
	transaction.Routes(transactionRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	ginTransactionRoute := ws.Group("/transaction")
	transactionRoute, _ := wrapper.NewRouterWrapper("transaction", ginTransactionRoute, getRoutesConfig())
	transaction.Routes(transactionRoute)
	return ws
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"transaction": {
				[]config.RouteConfig{
					{Name: "/send", Open: true},
					{Name: "/send-multiple", Open: true},
					{Name: "/cost", Open: true},
					{Name: "/:txhash", Open: true},
					{Name: "/:txhash/status", Open: true},
				},
			},
		},
	}
}
