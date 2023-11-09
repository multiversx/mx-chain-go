package gin

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/testscommon/api"
	"github.com/stretchr/testify/require"
)

func TestNewHttpServer(t *testing.T) {
	t.Parallel()

	t.Run("nil server should error", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(nil)
		require.Equal(t, apiErrors.ErrNilHttpServer, err)
		require.Nil(t, hs)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(&api.ServerStub{})
		require.NoError(t, err)
		require.NotNil(t, hs)
	})
}

func TestHttpServer_Start(t *testing.T) {
	t.Parallel()

	t.Run("server starts", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		serverStub := &api.ServerStub{
			ListenAndServeCalled: func() error {
				return nil
			},
			ShutdownCalled: func(ctx context.Context) error {
				wasCalled = true
				return nil
			},
		}
		hs, _ := NewHttpServer(serverStub)
		require.NotNil(t, hs)

		hs.Start()
		require.NoError(t, hs.Close())
		require.True(t, wasCalled)
	})
	t.Run("server is closed", func(t *testing.T) {
		t.Parallel()

		serverStub := &api.ServerStub{
			ListenAndServeCalled: func() error {
				return http.ErrServerClosed
			},
		}
		hs, _ := NewHttpServer(serverStub)
		require.NotNil(t, hs)

		hs.Start()
	})
	t.Run("server returns other error", func(t *testing.T) {
		t.Parallel()

		serverStub := &api.ServerStub{
			ListenAndServeCalled: func() error {
				return errors.New("other error")
			},
		}
		hs, _ := NewHttpServer(serverStub)
		require.NotNil(t, hs)

		hs.Start()
	})
}

func TestHttpServer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var hs *httpServer
	require.True(t, hs.IsInterfaceNil())

	hs, _ = NewHttpServer(&api.ServerStub{})
	require.False(t, hs.IsInterfaceNil())
}
