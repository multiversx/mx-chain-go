package sync

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMetaHeaderRequester(t *testing.T) {
	t.Parallel()

	headerRequester, err := NewMetaHeaderRequester(nil)
	require.Equal(t, process.ErrNilRequestHandler, err)
	require.Nil(t, headerRequester)

	headerRequester, err = NewMetaHeaderRequester(&testscommon.RequestHandlerStub{})
	require.Nil(t, err)
	require.False(t, headerRequester.IsInterfaceNil())
}

func TestMetaHeaderRequester_ShouldRequestHeader(t *testing.T) {
	t.Parallel()

	headerRequester, _ := NewMetaHeaderRequester(&testscommon.RequestHandlerStub{})
	require.False(t, headerRequester.IsInterfaceNil())

	require.False(t, headerRequester.ShouldRequestHeader(0))
	require.False(t, headerRequester.ShouldRequestHeader(1))
	require.False(t, headerRequester.ShouldRequestHeader(core.MainChainShardId))
	require.True(t, headerRequester.ShouldRequestHeader(core.MetachainShardId))
}

func TestMetaHeaderRequester_RequestHeader(t *testing.T) {
	t.Parallel()

	headerHash := []byte("hash")
	wasHeaderRequested := false
	requester := &testscommon.RequestHandlerStub{
		RequestMetaHeaderCalled: func(hash []byte) {
			require.Equal(t, headerHash, hash)
			wasHeaderRequested = true
		},
	}
	headerRequester, _ := NewMetaHeaderRequester(requester)
	headerRequester.RequestHeader(headerHash)
	require.True(t, wasHeaderRequested)
}
