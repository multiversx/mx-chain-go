package sync

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewExtendedHeaderRequester(t *testing.T) {
	t.Parallel()

	headerRequester, err := NewExtendedHeaderRequester(nil)
	require.Equal(t, process.ErrNilRequestHandler, err)
	require.Nil(t, headerRequester)

	headerRequester, err = NewExtendedHeaderRequester(&testscommon.ExtendedShardHeaderRequestHandlerStub{})
	require.Nil(t, err)
	require.False(t, headerRequester.IsInterfaceNil())
}

func TestExtendedHeaderRequester_ShouldRequestHeader(t *testing.T) {
	t.Parallel()

	headerRequester, _ := NewExtendedHeaderRequester(&testscommon.ExtendedShardHeaderRequestHandlerStub{})
	require.False(t, headerRequester.IsInterfaceNil())

	require.False(t, headerRequester.ShouldRequestHeader(0))
	require.False(t, headerRequester.ShouldRequestHeader(1))
	require.False(t, headerRequester.ShouldRequestHeader(core.MetachainShardId))
	require.True(t, headerRequester.ShouldRequestHeader(core.MainChainShardId))
}

func TestExtendedHeaderRequester_RequestHeader(t *testing.T) {
	t.Parallel()

	headerHash := []byte("hash")
	wasHeaderRequested := false
	requester := &testscommon.ExtendedShardHeaderRequestHandlerStub{
		RequestExtendedShardHeaderCalled: func(hash []byte) {
			require.Equal(t, headerHash, hash)
			wasHeaderRequested = true
		},
	}
	headerRequester, _ := NewExtendedHeaderRequester(requester)
	headerRequester.RequestHeader(headerHash)
	require.True(t, wasHeaderRequested)
}
