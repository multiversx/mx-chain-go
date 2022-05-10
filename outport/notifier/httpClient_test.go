package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/outport/notifier"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Hash string `json:"hash"`
}

func createMockHTTPClientArgs() notifier.HttpClientArgs {
	return notifier.HttpClientArgs{
		UseAuthorization: false,
		Username:         "user",
		Password:         "pass",
		BaseUrl:          "http://localhost:8080",
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	args := createMockHTTPClientArgs()
	client := notifier.NewHttpClient(args)
	require.NotNil(t, client)
}

func TestPOST(t *testing.T) {
	t.Parallel()

	wasCalled := false
	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wasCalled = true

		dataBytes, _ := json.Marshal(&testStruct{})
		_, err := w.Write(dataBytes)
		require.Nil(t, err)
	}))

	args := createMockHTTPClientArgs()
	args.BaseUrl = ws.URL

	client := notifier.NewHttpClient(args)
	require.NotNil(t, client)

	testPayload := testStruct{
		Hash: "hash1",
	}

	err := client.Post("/events/push", testPayload, nil)
	require.Nil(t, err)

	require.True(t, wasCalled)
}
