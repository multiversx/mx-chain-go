package libp2p_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerEventsDriver(t *testing.T) {
	t.Parallel()

	driver := libp2p.NewPeerEventsDriver()
	assert.False(t, check.IfNil(driver))
}

func TestPeerEventsDriver_AddHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil handler should error", func(t *testing.T) {
		driver := libp2p.NewPeerEventsDriver()

		err := driver.AddHandler(nil)
		assert.Equal(t, p2p.ErrNilPeerEventsHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		driver := libp2p.NewPeerEventsDriver()
		handler := &mock.PeerEventsHandlerStub{}

		err := driver.AddHandler(handler)
		assert.Nil(t, err)

		assert.Equal(t, 1, len(driver.GetHandlers()))
		assert.True(t, driver.GetHandlers()[0] == handler) // pointer testing

		t.Run("second element added should be second in the list", func(t *testing.T) {
			handler2 := &mock.PeerEventsHandlerStub{}

			err = driver.AddHandler(handler2)
			assert.Nil(t, err)

			assert.Equal(t, 2, len(driver.GetHandlers()))
			assert.True(t, driver.GetHandlers()[0] == handler)  // pointer testing
			assert.True(t, driver.GetHandlers()[1] == handler2) // pointer testing
		})
	})
}

func TestConnectionMonitorWrapper_Connected(t *testing.T) {
	t.Parallel()

	testPid := core.PeerID("pid")
	testConnection := "connection"
	t.Run("no handler added should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()

		driver := libp2p.NewPeerEventsDriver()
		driver.Connected(testPid, testConnection)
	})
	t.Run("2 handlers should work", func(t *testing.T) {
		driver := libp2p.NewPeerEventsDriver()

		handler1Called := false
		handler1 := &mock.PeerEventsHandlerStub{
			ConnectedCalled: func(pid core.PeerID, connection string) {
				assert.Equal(t, testPid, pid)
				assert.Equal(t, testConnection, connection)
				handler1Called = true
			},
		}

		handler2Called := false
		handler2 := &mock.PeerEventsHandlerStub{
			ConnectedCalled: func(pid core.PeerID, connection string) {
				assert.Equal(t, testPid, pid)
				assert.Equal(t, testConnection, connection)
				handler2Called = true
			},
		}

		_ = driver.AddHandler(handler1)
		_ = driver.AddHandler(handler2)

		driver.Connected(testPid, testConnection)

		assert.True(t, handler1Called)
		assert.True(t, handler2Called)
	})
}

func TestConnectionMonitorWrapper_Disconnected(t *testing.T) {
	t.Parallel()

	testPid := core.PeerID("pid")
	t.Run("no handler added should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()

		driver := libp2p.NewPeerEventsDriver()
		driver.Disconnected(testPid)
	})
	t.Run("2 handlers should work", func(t *testing.T) {
		driver := libp2p.NewPeerEventsDriver()

		handler1Called := false
		handler1 := &mock.PeerEventsHandlerStub{
			DisconnectedCalled: func(pid core.PeerID) {
				assert.Equal(t, testPid, pid)
				handler1Called = true
			},
		}

		handler2Called := false
		handler2 := &mock.PeerEventsHandlerStub{
			DisconnectedCalled: func(pid core.PeerID) {
				assert.Equal(t, testPid, pid)
				handler2Called = true
			},
		}

		_ = driver.AddHandler(handler1)
		_ = driver.AddHandler(handler2)

		driver.Disconnected(testPid)

		assert.True(t, handler1Called)
		assert.True(t, handler2Called)
	})
}
