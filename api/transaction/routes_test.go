package transaction_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction/mock"
	tr "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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

//------- GenerateTransaction

func TestGenerateTransaction_WithParametersShouldReturnTransaction(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	facade := mock.Facade{
		GenerateTransactionHandler: func(sender string, receiver string, value big.Int, code string) (transaction *tr.Transaction, e error) {
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
	assert.Contains(t, transactionResponse.Error, "Validation error: ")
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
	assert.Contains(t, transactionResponse.Error, "Invalid app context")
}

func TestGenerateTransaction_GenerationFailedShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	facade := mock.Facade{
		GenerateTransactionHandler: func(sender string, receiver string, value big.Int, code string) (transaction *tr.Transaction, e error) {
			return nil, errors.New("generation failed")
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
	assert.Contains(t, transactionResponse.Error, "generation failed")
}

//------- GenerateAndSendMultipleTransaction

func TestGenerateAndSendMultipleTransaction_WithParametersShouldReturnNoError(t *testing.T) {
	t.Parallel()
	receiver := "multipleReceiver"
	value := big.NewInt(5)
	noTxs := 10

	facade := mock.Facade{
		GenerateAndSendBulkTransactionsHandler: func(receiver string, value big.Int,
			noTxs uint64) error {
			return nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"receiver":"%s",`+
			`"value":%s,`+
			`"noTxs":%d}`, receiver, value, noTxs)

	req, _ := http.NewRequest("POST", "/transaction/generateAndSendMultiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	multipleTransactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &multipleTransactionResponse)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", multipleTransactionResponse.Error)
	assert.Equal(t, fmt.Sprintf("%d", noTxs), multipleTransactionResponse.Message)
}

func TestGenerateAndSendMultipleTransaction_WithBadJsonShouldReturnBadRequest(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generateAndSendMultiple", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Validation error: ")
}

func TestGenerateAndSendMultipleTransaction_WithBadJsonShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/generateAndSendMultiple", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Invalid app context")
}

func TestGenerateAndSendMultipleTransaction_GenerationFailedShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()
	receiver := "multipleReceiver"
	value := big.NewInt(5)
	noTxs := 10

	facade := mock.Facade{
		GenerateAndSendBulkTransactionsHandler: func(receiver string, value big.Int,
			noTxs uint64) error {
			return errors.New("generation failed")
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"receiver":"%s",`+
			`"value":%s,`+
			`"noTxs":%d}`, receiver, value, noTxs)

	req, _ := http.NewRequest("POST", "/transaction/generateAndSendMultiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	multipleTransactionResponse := GeneralResponse{}
	loadResponse(resp.Body, &multipleTransactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, multipleTransactionResponse.Error, "generation failed")
}

//------- GetTransaction

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
				Value:   *value,
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
				Value:   *value,
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

func TestGetTransaction_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/transaction/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, transactionResponse.Error, "Invalid app context")
}

func TestGetTransaction_GetFailedShouldReturnInternalServerError(t *testing.T) {
	hash := "hash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.Transaction, e error) {
			return nil, errors.New("get failed")
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+hash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Transaction getting failed")
}

//------- SendTransaction

func TestSendTransaction_WithParametersShouldReturnTransaction(t *testing.T) {
	t.Parallel()

	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	sig := hex.EncodeToString([]byte("sig"))

	facade := mock.Facade{
		SendTransactionHandler: func(nonce uint64, sender string, receiver string, value big.Int, code string, signature []byte) (*tr.Transaction, error) {
			return &tr.Transaction{
				SndAddr:   []byte(sender),
				RcvAddr:   []byte(receiver),
				Data:      []byte(code),
				Value:     value,
				Signature: signature,
			}, nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s",`+
			`"signature":"%s"}`, sender, receiver, value, data, sig)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

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
	assert.Equal(t, sig, txResp.Signature)
}

func TestSendTransaction_WithBadJsonShouldReturnBadRequest(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Validation error: ")
}

func TestSendTransaction_WithBadJsonShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Invalid app context")
}

func TestSendTransaction_WithBadSignatureShouldReturnBadRequest(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	sig := []byte("sig is not hex!")

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s",`+
			`"signature":"%s"}`, sender, receiver, value, data, sig)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Invalid signature, could not decode hex value")
}

func TestSendTransaction_SendFailedShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	sig := hex.EncodeToString([]byte("sig"))

	facade := mock.Facade{
		SendTransactionHandler: func(nonce uint64, sender string, receiver string, value big.Int, code string, signature []byte) (*tr.Transaction, error) {
			return nil, errors.New("send failed")
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s",`+
			`"signature":"%s"}`, sender, receiver, value, data, sig)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, "send failed")
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
	gin.SetMode(gin.TestMode)
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
	gin.SetMode(gin.TestMode)
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	transactionRoute := ws.Group("/transaction")
	transaction.Routes(transactionRoute)
	return ws
}
