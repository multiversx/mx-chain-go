package softwareVersion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStableTagProvider(t *testing.T) {
	t.Parallel()

	stp := NewStableTagProvider("location")
	assert.NotNil(t, stp)
}

func TestStableTagProvider_FetchTagVersion(t *testing.T) {
	t.Parallel()

	t.Run("get failure should error", func(t *testing.T) {
		t.Parallel()

		stp := NewStableTagProvider("invalid location")
		version, err := stp.FetchTagVersion()
		assert.Error(t, err)
		assert.Empty(t, version)
	})
	t.Run("unmarshal failure should error", func(t *testing.T) {
		t.Parallel()

		handlerWasCalled := false
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerWasCalled = true
			_, err := w.Write([]byte("invalid data to unmarshal"))
			assert.NoError(t, err)
		}))
		defer ts.Close()

		stp := NewStableTagProvider(ts.URL)
		version, err := stp.FetchTagVersion()
		assert.Error(t, err)
		assert.Empty(t, version)

		assert.True(t, handlerWasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedTagVersion := &tagVersion{
			TagVersion: "1.0.0",
		}
		providedTagBytes, _ := json.Marshal(providedTagVersion)
		handlerWasCalled := false
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerWasCalled = true
			_, err := w.Write(providedTagBytes)
			assert.NoError(t, err)
		}))
		defer ts.Close()

		stp := NewStableTagProvider(ts.URL)
		version, err := stp.FetchTagVersion()
		assert.NoError(t, err)
		assert.Equal(t, providedTagVersion.TagVersion, version)

		assert.True(t, handlerWasCalled)
	})
}

func TestStableTagProvider_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var stp *stableTagProvider
	assert.True(t, stp.IsInterfaceNil())

	stp = NewStableTagProvider("location")
	assert.False(t, stp.IsInterfaceNil())
}
