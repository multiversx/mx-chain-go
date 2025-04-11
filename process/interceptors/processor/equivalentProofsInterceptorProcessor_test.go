package processor

import (
	"testing"

	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
)

func createMockArgEquivalentProofsInterceptorProcessor() ArgEquivalentProofsInterceptorProcessor {
	return ArgEquivalentProofsInterceptorProcessor{
		EquivalentProofsPool: &dataRetriever.ProofsPoolMock{},
		Marshaller:           &marshallerMock.MarshalizerMock{},
		PeerShardMapper:      &processMocks.PeerShardMapperStub{},
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
	}
}

func TestEquivalentProofsInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var epip *equivalentProofsInterceptorProcessor
	require.True(t, epip.IsInterfaceNil())

	epip, _ = NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.False(t, epip.IsInterfaceNil())
}

func TestNewEquivalentProofsInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil EquivalentProofsPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.EquivalentProofsPool = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilProofsPool, err)
		require.Nil(t, epip)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.Marshaller = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
		require.Nil(t, epip)
	})
	t.Run("nil PeerShardMapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.PeerShardMapper = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilPeerShardMapper, err)
		require.Nil(t, epip)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsInterceptorProcessor()
		args.NodesCoordinator = nil

		epip, err := NewEquivalentProofsInterceptorProcessor(args)
		require.Equal(t, process.ErrNilNodesCoordinator, err)
		require.Nil(t, epip)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
		require.NoError(t, err)
		require.NotNil(t, epip)
	})
}

func TestEquivalentProofsInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.NoError(t, err)

	// coverage only
	require.Nil(t, epip.Validate(nil, ""))
}

func TestEquivalentProofsInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.NoError(t, err)

	// coverage only
	err = epip.Save(nil, "", "")
	require.Nil(t, err)
}

func TestEquivalentProofsInterceptorProcessor_RegisterHandler(t *testing.T) {
	t.Parallel()

	epip, err := NewEquivalentProofsInterceptorProcessor(createMockArgEquivalentProofsInterceptorProcessor())
	require.NoError(t, err)

	// coverage only
	epip.RegisterHandler(nil)
}
