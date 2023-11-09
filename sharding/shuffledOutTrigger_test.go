package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/stretchr/testify/require"
)

func TestNewShuffledOutTrigger_NilOwnPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	endProcessHandler := func(_ endProcess.ArgEndProcess) error {
		return nil
	}

	sot, err := NewShuffledOutTrigger(nil, 0, endProcessHandler)
	require.True(t, check.IfNil(sot))
	require.Equal(t, ErrNilOwnPublicKey, err)
}

func TestNewShuffledOutTrigger_NilEndOfProcessingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sot, err := NewShuffledOutTrigger([]byte("pk"), 0, nil)
	require.True(t, check.IfNil(sot))
	require.Equal(t, ErrNilEndOfProcessingHandler, err)
}

func TestNewShuffledOutTrigger_ShouldWork(t *testing.T) {
	t.Parallel()

	handler := func(_ endProcess.ArgEndProcess) error {
		return nil
	}

	sot, err := NewShuffledOutTrigger([]byte("opk"), 0, handler)
	require.False(t, check.IfNil(sot))
	require.NoError(t, err)
}

func TestShuffledOutTrigger_Process_SameShardIDShouldReturn(t *testing.T) {
	t.Parallel()

	endProcessHandlerWasCalled := false

	endProcessHandler := func(_ endProcess.ArgEndProcess) error {
		endProcessHandlerWasCalled = true
		return nil
	}
	shardID := uint32(1)

	sot, _ := NewShuffledOutTrigger([]byte("opk"), shardID, endProcessHandler)
	require.False(t, check.IfNil(sot))

	err := sot.Process(shardID)
	require.NoError(t, err)
	require.False(t, endProcessHandlerWasCalled)
}

func TestShuffledOutTrigger_Process_DifferentShardIDShouldCallEndProcessHandler(t *testing.T) {
	t.Parallel()

	endProcessHandlerWasCalled := false

	endProcessHandler := func(_ endProcess.ArgEndProcess) error {
		endProcessHandlerWasCalled = true
		return nil
	}
	shardID := uint32(1)
	newShardID := shardID + 1

	sot, _ := NewShuffledOutTrigger([]byte("opk"), shardID, endProcessHandler)
	require.False(t, check.IfNil(sot))

	err := sot.Process(newShardID)
	require.NoError(t, err)
	require.True(t, endProcessHandlerWasCalled)
	require.Equal(t, newShardID, sot.CurrentShardID())
}

func TestShuffledOutTrigger_DifferentShardIDShouldCallRegisteredHandlers(t *testing.T) {
	t.Parallel()

	handler1WasCalled, handler2WasCalled := false, false

	endProcessHandler := func(_ endProcess.ArgEndProcess) error {
		return nil
	}
	shardID := uint32(1)
	newShardID := shardID + 1

	handler1 := func(_ uint32) {
		handler1WasCalled = true
	}
	handler2 := func(_ uint32) {
		handler2WasCalled = true
	}

	sot, _ := NewShuffledOutTrigger([]byte("opk"), shardID, endProcessHandler)
	require.False(t, check.IfNil(sot))

	sot.RegisterHandler(handler1)
	sot.RegisterHandler(handler2)

	err := sot.Process(newShardID)
	require.NoError(t, err)

	require.True(t, handler1WasCalled)
	require.True(t, handler2WasCalled)
}
