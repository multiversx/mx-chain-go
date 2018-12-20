package transaction_test

import (
	"bytes"
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
	"github.com/stretchr/testify/assert"
)

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type TransactionResponse struct {
	GeneralResponse
	TxResp transaction.TxResponse `json:"transaction"`
}

func TestGenerateTransaction_WithParameters_ShouldReturnTransaction(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	facade := mock.Facade{
		GenerateTransactionHandler: func(sender string, receiver string, amount big.Int, code string) (transaction *tr.Transaction, e error) {
			return &tr.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(code),
				//TODO: change this to big.Int
				Value: uint64(amount.Int64()),
			}, nil
		},
	}

	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"sender":"%s",`+
			`"receiver":"%s",`+
			`"value":%s,`+
			`"data":"%s"}`, sender, receiver, value, data)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	txResp := transactionResponse.TxResp

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", transactionResponse.Error)
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, data, txResp.Data)
}

func TestGetTransaction_WithCorrectHash_ShouldReturnTransaction(t *testing.T) {
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
				//TODO: change this to big.Int
				Value: uint64(value.Int64()),
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
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, data, txResp.Data)
}

func TestGetTransaction_WithUnknownHash_ShouldReturnNil(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	hs := "hash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.Transaction, e error) {
			if hash != hs {
				return nil, nil
			}
			return &tr.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(data),
				//TODO: change this to big.Int
				Value: uint64(value.Int64()),
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+hs, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	txResp := transactionResponse.TxResp

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, data, txResp.Data)
}

func TestGenerateTransaction_WithBadJsonShouldReturnBadRequest(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}

	ws := startNodeServer(&facade)

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Validation error: ")
}

func TestGetTransactionFailsWithWrongFacadeTypeConversion(t *testing.T) {
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

func TestGenerateTransaction_WithBadJsonShouldReturnInternalServerError(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	badJsonString := "bad"

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer([]byte(badJsonString)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	transactionResponse := TransactionResponse{}
	loadResponse(resp.Body, &transactionResponse)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, transactionResponse.Error, "Invalid app context")
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

func startNodeServer(handler transaction.Handler) *gin.Engine {
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
