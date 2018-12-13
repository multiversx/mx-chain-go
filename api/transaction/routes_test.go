package transaction_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/transaction/mock"
	tr "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GeneralResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type TransactionResponse struct {
	GeneralResponse
	TxResp transaction.TxResponse `json:"transaction"`
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

func TestGenerateTransaction_WithParameters_ShouldReturnTransaction(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"

	facade := mock.Facade{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (transaction *tr.Transaction, e error) {
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

	jsonStr := fmt.Sprintf("{\"sender\":\"%s\","+
		"\"receiver\":\"%s\","+
		"\"value\":%s,"+
		"\"data\":\"%s\"}", sender, receiver, value, data)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer([]byte(jsonStr)))

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

func TestGetTransaction_WithCorrectHash_ShouldReturnTransaction(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	hash := "hash"
	facade := mock.Facade{
		GetTransactionCalled: func(hash string) (i *tr.Transaction, e error) {
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
		GetTransactionCalled: func(hash string) (i *tr.Transaction, e error) {
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
