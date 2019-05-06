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

	errors2 "github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction/mock"
	tr "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
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

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateTransaction_WithParametersShouldReturnTransaction(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	facade := mock.Facade{
		GenerateTransactionHandler: func(sender string, receiver string, value *big.Int, code string) (transaction *tr.Transaction, e error) {
			return &tr.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(code),
				Value:   value,
			}, nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s"}`, sender, receiver, value, data)

	req, _ := http.NewRequest("POST", "/transaction/generate", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	txResp := transactionResponse.TxResp

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", transactionResponse.Error)
	assert.Equal(t, hex.EncodeToString([]byte(sender)), txResp.Sender)
	assert.Equal(t, hex.EncodeToString([]byte(receiver)), txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, data, txResp.Data)
}

func TestGenerateAndSendMultipleTransaction_WithParametersShouldReturnNoError(t *testing.T) {
	t.Parallel()
	receiver := "multipleReceiver"
	value := big.NewInt(5)
	txCount := 10

	facade := mock.Facade{
		GenerateAndSendBulkTransactionsHandler: func(receiver string, value *big.Int,
			txCount uint64) error {
			return nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"receiver":"%s",`+
			`"value":%s,`+
			`"txCount":%d}`, receiver, value, txCount)

	req, _ := http.NewRequest("POST", "/transaction/generate-and-send-multiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	multipleTransactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &multipleTransactionResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", multipleTransactionResponse.Error)
	assert.Equal(t, fmt.Sprintf("%d", txCount), multipleTransactionResponse.Message)
}

func TestGenerateAndSendMultipleTransactionOneByOne_WithParametersShouldReturnNoError(t *testing.T) {
	t.Parallel()
	receiver := "multipleReceiver"
	value := big.NewInt(5)
	txCount := 10

	facade := mock.Facade{
		GenerateAndSendBulkTransactionsOneByOneHandler: func(receiver string, value *big.Int,
			txCount uint64) error {
			return nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"receiver":"%s",`+
			`"value":%s,`+
			`"txCount":%d}`, receiver, value, txCount)

	req, _ := http.NewRequest("POST", "/transaction/generate-and-send-multiple-one-by-one", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	multipleTransactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &multipleTransactionResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", multipleTransactionResponse.Error)
	assert.Equal(t, fmt.Sprintf("%d", txCount), multipleTransactionResponse.Message)
}

func TestGetTransaction_WithCorrectHashShouldReturnTransaction(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	hash := "hash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.Transaction, e error) {
			return &tr.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(data),
				Value:   value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+hash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	txResp := transactionResponse.TxResp

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, hex.EncodeToString([]byte(sender)), txResp.Sender)
	assert.Equal(t, hex.EncodeToString([]byte(receiver)), txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, data, txResp.Data)
}

func TestGetTransaction_WithUnknownHashShouldReturnNil(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	hs := "hash"
	wrongHash := "wronghash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.Transaction, e error) {
			if hash != hs {
				return nil, nil
			}
			return &tr.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(data),
				Value:   value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+wrongHash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Nil(t, transactionResponse.TxResp)
}

func TestGenerateTransaction_WithBadJsonShouldReturnBadRequest(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generate", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrValidation.Error())
}

func TestGenerateAndSendMultipleTransaction_WithBadJsonShouldReturnBadRequest(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generate-and-send-multiple", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrValidation.Error())
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

func TestGenerateTransaction_WithBadJsonShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generate", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrInvalidAppContext.Error())
}

func TestGenerateAndSendMultipleTransaction_WithBadJsonShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generate-and-send-multiple", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrInvalidAppContext.Error())
}

func TestGenerateTransaction_ErrorsWhenFacadeGenerateTransactionFails(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	errorString := "generate transaction error"
	facade := mock.Facade{
		GenerateTransactionHandler: func(sender string, receiver string, value *big.Int, code string) (transaction *tr.Transaction, e error) {
			return nil, errors.New(errorString)
		},
	}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s"}`, sender, receiver, value, data)

	req, _ := http.NewRequest("POST", "/transaction/generate", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, fmt.Sprintf("%s: %s", errors2.ErrTxGenerationFailed.Error(), errorString), transactionResponse.Error)
	assert.Empty(t, transactionResponse.TxResp)
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

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s"}`, sender, receiver, value, data)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrValidation.Error())
	assert.Empty(t, transactionResponse.TxResp)
}

func TestSendTransaction_InvalidHexSignatureShouldError(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "not#only$hex%characters^"

	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"signature":"%s",`+
			`"data":"%s"}`, sender, receiver, value, signature, data)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, errors2.ErrInvalidSignatureHex.Error())
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
		SendTransactionHandler: func(nonce uint64, sender string, receiver string, value *big.Int,
			code string, signature []byte) (transaction *tr.Transaction, e error) {
			return nil, errors.New(errorString)
		},
	}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"signature":"%s",`+
			`"data":"%s"}`, sender, receiver, value, signature, data)

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

	facade := mock.Facade{
		SendTransactionHandler: func(nonce uint64, sender string, receiver string, value *big.Int,
			code string, signature []byte) (transaction *tr.Transaction, e error) {
			return &tr.Transaction{
				Nonce:     nonce,
				SndAddr:   []byte(sender),
				RcvAddr:   []byte(receiver),
				Value:     value,
				Data:      []byte(code),
				Signature: signature,
			}, nil
		},
	}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(`{
		"nonce": %d,
		"sender": "%s",
		"receiver": "%s",
		"value": %s,
		"signature": "%s",
		"data": "%s"
	}`, nonce, sender, receiver, value, signature, data)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, transactionResponse.Error)
	assert.Equal(t, transactionResponse.TxResp.Nonce, nonce)
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
	transactionRoute := ws.Group("/transaction")
	if handler != nil {
		transactionRoute.Use(middleware.WithElrondFacade(handler))
	}
	transaction.Routes(transactionRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	transactionRoute := ws.Group("/transaction")
	transaction.Routes(transactionRoute)
	return ws
}
