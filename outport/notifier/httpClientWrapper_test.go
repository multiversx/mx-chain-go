package notifier_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/outport/notifier"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Hash string `json:"hash"`
}

func createMockHTTPClientArgs() notifier.HTTPClientWrapperArgs {
	return notifier.HTTPClientWrapperArgs{
		UseAuthorization:  false,
		Username:          "user",
		Password:          "pass",
		BaseUrl:           "http://localhost:8080",
		RequestTimeoutSec: 60,
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("invalid request timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockHTTPClientArgs()
		args.RequestTimeoutSec = 0
		client, err := notifier.NewHTTPWrapperClient(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, notifier.ErrInvalidValue))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHTTPClientArgs()
		client, err := notifier.NewHTTPWrapperClient(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(client))
	})
}

func TestPOST(t *testing.T) {
	t.Parallel()

	testPayload := testStruct{
		Hash: "hash1",
	}
	dataBytes, _ := json.Marshal(testPayload)

	wasCalled := false
	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wasCalled = true

		_, err := w.Write(dataBytes)
		require.Nil(t, err)
	}))

	args := createMockHTTPClientArgs()
	args.BaseUrl = ws.URL

	client, err := notifier.NewHTTPWrapperClient(args)
	require.Nil(t, err)
	require.NotNil(t, client)

	err = client.Post("/events/push", testPayload)
	require.Nil(t, err)

	require.True(t, wasCalled)
}

func TestPOSTShouldFail(t *testing.T) {
	t.Parallel()

	testPayload := testStruct{
		Hash: "hash1",
	}
	dataBytes, _ := json.Marshal(testPayload)

	statusCode := http.StatusBadGateway

	wasCalled := false
	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wasCalled = true

		w.WriteHeader(statusCode)
		_, err := w.Write(dataBytes)
		require.Nil(t, err)
	}))

	args := createMockHTTPClientArgs()
	args.BaseUrl = ws.URL

	client, err := notifier.NewHTTPWrapperClient(args)
	require.Nil(t, err)
	require.NotNil(t, client)

	err = client.Post("/events/push", testPayload)
	require.True(t, strings.Contains(err.Error(), http.StatusText(statusCode)))

	require.True(t, wasCalled)
}
