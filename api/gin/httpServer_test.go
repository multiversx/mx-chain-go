package gin

import (
	"fmt"
	"testing"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/testscommon/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHttpServer(t *testing.T) {
	t.Parallel()

	t.Run("nil engine should error", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer("localhost:0", "tcp4", nil)
		require.Equal(t, apiErrors.ErrNilHTTPHandler, err)
		require.Nil(t, hs)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer("localhost:0", "tcp4", &api.HandlerStub{})
		require.NoError(t, err)
		require.NotNil(t, hs)
	})
}

func TestHttpServer_PortAndAddressShouldWork(t *testing.T) {
	t.Parallel()

	localHostTCP4IP := "127.0.0.1"

	t.Run("using localhost as interface", func(t *testing.T) {
		hs, _ := NewHttpServer("localhost:0", "tcp4", &api.HandlerStub{})
		boundPort := hs.Port()
		assert.Greater(t, boundPort, 0) // no port 0 accepted
		assert.Equal(t, fmt.Sprintf("%s:%d", localHostTCP4IP, boundPort), hs.Address())
	})
	t.Run("using 127.0.0.1 as interface", func(t *testing.T) {
		hs, _ := NewHttpServer(localHostTCP4IP+":0", "tcp4", &api.HandlerStub{})
		boundPort := hs.Port()
		assert.Greater(t, boundPort, 0) // no port 0 accepted
		assert.Equal(t, fmt.Sprintf("%s:%d", localHostTCP4IP, boundPort), hs.Address())
	})
	// TODO
	//t.Run("using empty string as interface", func(t *testing.T) {
	//	hs, _ := NewHttpServer(":0", "tcp4", &api.HandlerStub{})
	//	boundPort := hs.Port()
	//	assert.Greater(t, boundPort, 0) // no port 0 accepted
	//	assert.Equal(t, fmt.Sprintf(":%d", boundPort), hs.Address())
	//})
}

func TestHttpServer_StartAndServe(t *testing.T) {
	t.Parallel()

	// TODO: more connection tests should be added here. Along with different networks like tcp, tcp4 and tcp6. Ensure that
	//  a client can connect to :<port> when defining all networks.
}

//
//func TestHttpServer_Start(t *testing.T) {
//	t.Parallel()
//
//	t.Run("server starts", func(t *testing.T) {
//		t.Parallel()
//p6
//		wasCalled := false
//		serverStub := &api.ServerStub{
//			ListenAndServeCalled: func() error {
//				return nil
//			},
//			ShutdownCalled: func(ctx context.Context) error {
//				wasCalled = true
//				return nil
//			},
//		}
//		hs, _ := NewHttpServer(serverStub)
//		require.NotNil(t, hs)
//
//		hs.Start()
//		require.NoError(t, hs.Close())
//		require.True(t, wasCalled)
//	})
//	t.Run("server is closed", func(t *testing.T) {
//		t.Parallel()
//
//		serverStub := &api.ServerStub{
//			ListenAndServeCalled: func() error {
//				return http.ErrServerClosed
//			},
//		}
//		hs, _ := NewHttpServer(serverStub)
//		require.NotNil(t, hs)
//
//		hs.Start()
//	})
//	t.Run("server returns other error", func(t *testing.T) {
//		t.Parallel()
//
//		serverStub := &api.ServerStub{
//			ListenAndServeCalled: func() error {
//				return errors.New("other error")
//			},
//		}
//		hs, _ := NewHttpServer(serverStub)
//		require.NotNil(t, hs)
//
//		hs.Start()
//	})
//}
//
//func TestHttpServer_IsInterfaceNil(t *testing.T) {
//	t.Parallel()
//
//	var hs *httpServer
//	require.True(t, hs.IsInterfaceNil())
//
//	hs, _ = NewHttpServer(&api.ServerStub{})
//	require.False(t, hs.IsInterfaceNil())
//}
