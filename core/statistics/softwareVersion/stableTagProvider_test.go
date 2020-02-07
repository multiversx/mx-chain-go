package softwareVersion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewStableTagProvider(t *testing.T) {
	t.Parallel()

	stp := NewStableTagProvider("location")
	assert.False(t, check.IfNil(stp))
}

func TestStableTagProvider_FetchTagVersion(t *testing.T) {
	t.Parallel()

	handlerWasCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerWasCalled = true
	}))
	defer ts.Close()

	stp := NewStableTagProvider(ts.URL)
	_, _ = stp.FetchTagVersion()

	assert.True(t, handlerWasCalled)
}
