package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"
)

func TestAppStatusRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/appstatus", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestBalanceRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/balance", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCheckFreePortRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/checkfreeport", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestExitRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/exit", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGenerateKeyRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/generatepublickeyandprivateKey", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGetBlockFromHashRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/getblockfromhash", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGetNextPrivateKeyRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/getNextPrivateKey", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGetShardRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/getprivatepublickeyshard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGetStatsRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/getStats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestGetTransactionFromHashRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/gettransactionfromhash", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestPingRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestReceiptRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/receipt", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestSendRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/send", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
func TestSendMultipleTransactionsRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/sendMultipleTransactions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestSendMultipleTransactionsToAllShardsRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/sendMultipleTransactionsToAllShards", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestShardOfAddressRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/shardofaddress", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestStartRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/start", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestStatusRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestStopRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/node/stop", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
