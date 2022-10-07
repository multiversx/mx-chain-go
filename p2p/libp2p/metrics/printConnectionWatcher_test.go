package metrics

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewPrintConnectionsWatcher(t *testing.T) {
	t.Parallel()

	t.Run("invalid value for time to live parameter should error", func(t *testing.T) {
		t.Parallel()

		pcw, err := NewPrintConnectionsWatcher(minTimeToLive - time.Nanosecond)
		assert.True(t, check.IfNil(pcw))
		assert.True(t, errors.Is(err, errInvalidValueForTimeToLiveParam))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pcw, err := NewPrintConnectionsWatcher(minTimeToLive)
		assert.False(t, check.IfNil(pcw))
		assert.Nil(t, err)

		_ = pcw.Close()
	})
}

func TestPrintConnectionsWatcher_Close(t *testing.T) {
	t.Parallel()

	t.Run("no iteration has been done", func(t *testing.T) {
		t.Parallel()

		pcw, _ := NewPrintConnectionsWatcher(time.Hour)
		err := pcw.Close()

		assert.Nil(t, err)
		time.Sleep(time.Second) // allow the go routine to close
		assert.True(t, pcw.goRoutineClosed.IsSet())
	})
	t.Run("iterations were done", func(t *testing.T) {
		t.Parallel()

		pcw, _ := NewPrintConnectionsWatcher(time.Second)
		time.Sleep(time.Second * 4)
		err := pcw.Close()

		assert.Nil(t, err)
		time.Sleep(time.Second) // allow the go routine to close
		assert.True(t, pcw.goRoutineClosed.IsSet())
	})

}

func TestPrintConnectionsWatcher_Connected(t *testing.T) {
	t.Parallel()

	t.Run("invalid connection", func(t *testing.T) {
		providedPid := core.PeerID("pid")
		connection := " "
		numCalled := 0

		handler := func(pid core.PeerID, conn string) {
			numCalled++
		}
		pcw, _ := NewPrintConnectionsWatcherWithHandler(time.Hour, handler)

		pcw.Connected(providedPid, connection)
		assert.Equal(t, 0, numCalled)
	})
	t.Run("valid connection", func(t *testing.T) {
		providedPid := core.PeerID("pid")
		connection := "connection"
		numCalled := 0

		handler := func(pid core.PeerID, conn string) {
			numCalled++
			assert.Equal(t, providedPid, pid)
			assert.Equal(t, connection, conn)
		}
		pcw, _ := NewPrintConnectionsWatcherWithHandler(time.Hour, handler)

		pcw.Connected(providedPid, connection)
		assert.Equal(t, 1, numCalled)
		pcw.Connected(providedPid, connection)
		assert.Equal(t, 1, numCalled)
	})
}

func TestPrintConnectionsWatcher_DisconnectedShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	pcw, _ := NewPrintConnectionsWatcher(time.Second)
	pcw.Disconnected("")

	_ = pcw.Close()
}

func TestLogPrintHandler_shouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	logPrintHandler("pid", "connection")
}
