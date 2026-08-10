package processor_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func TestNewUniqueChunksProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil cache should error", func(t *testing.T) {
		t.Parallel()

		ucp, err := processor.NewUniqueChunksProcessor(nil, &mock.MarshalizerMock{}, &mock.HasherStub{})
		require.True(t, check.IfNil(ucp))
		require.Equal(t, process.ErrNilInterceptedDataCache, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		ucp, err := processor.NewUniqueChunksProcessor(&cache.CacherStub{}, nil, &mock.HasherStub{})
		require.True(t, check.IfNil(ucp))
		require.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		ucp, err := processor.NewUniqueChunksProcessor(&cache.CacherStub{}, &mock.MarshalizerMock{}, nil)
		require.True(t, check.IfNil(ucp))
		require.Equal(t, process.ErrNilHasher, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ucp, err := processor.NewUniqueChunksProcessor(&cache.CacherStub{}, &mock.MarshalizerMock{}, &mock.HasherStub{})
		require.False(t, check.IfNil(ucp))
		require.NoError(t, err)

		err = ucp.Close() // coverage only
		require.NoError(t, err)
	})
}

func TestUniqueChunksProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ucp, _ := processor.NewUniqueChunksProcessor(nil, &mock.MarshalizerMock{}, &mock.HasherStub{})
	require.True(t, ucp.IsInterfaceNil())

	ucp, _ = processor.NewUniqueChunksProcessor(&cache.CacherStub{}, &mock.MarshalizerMock{}, &mock.HasherStub{})
	require.False(t, ucp.IsInterfaceNil())
}

func TestUniqueChunksProcessor_CheckBatch(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should return empty result", func(t *testing.T) {
		t.Parallel()

		ucp, _ := processor.NewUniqueChunksProcessor(&cache.CacherStub{}, &mock.MarshalizerMock{}, &mock.HasherStub{})
		result, err := ucp.CheckBatch(nil, nil, p2p.Broadcast)
		require.Equal(t, process.CheckedChunkResult{}, result)
		require.NoError(t, err)
	})
	t.Run("marshaling error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("marshal error")
		marshaller := &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		ucp, _ := processor.NewUniqueChunksProcessor(&cache.CacherStub{}, marshaller, &mock.HasherStub{})
		result, err := ucp.CheckBatch(&batch.Batch{Data: [][]byte{{1, 2, 3}}}, nil, p2p.Broadcast)
		require.Equal(t, process.CheckedChunkResult{}, result)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		b := &batch.Batch{Data: [][]byte{{1, 2, 3}}}
		hasher := &hashingMocks.HasherMock{}
		marshaller := &mock.MarshalizerMock{}
		batchHash, _ := core.CalculateHash(marshaller, hasher, b)

		cacheMock := cache.NewCacherMock()
		ucp, _ := processor.NewUniqueChunksProcessor(cacheMock, marshaller, hasher)

		// First check should succeed
		result, err := ucp.CheckBatch(b, nil, p2p.Broadcast)
		require.Equal(t, process.CheckedChunkResult{}, result)
		require.NoError(t, err)

		ucp.MarkVerified(b, p2p.Broadcast)

		// Verify it was added to cache
		_, ok := cacheMock.Get(batchHash)
		require.True(t, ok)

		// Second check with same batch should fail
		result, err = ucp.CheckBatch(b, nil, p2p.Broadcast)
		require.Equal(t, process.CheckedChunkResult{}, result)
		require.Equal(t, process.ErrDuplicatedInterceptedDataNotAllowed, err)
	})
}

func TestUniqueChunksProcessor_MarkVerified(t *testing.T) {
	t.Parallel()

	b := &batch.Batch{Data: [][]byte{{1, 2, 3}}}
	hasher := &hashingMocks.HasherMock{}
	marshaller := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("marshal error")
		},
	}
	cacheMock := cache.NewCacherMock()
	ucp, _ := processor.NewUniqueChunksProcessor(cacheMock, marshaller, hasher)

	// nil batch, early exit
	ucp.MarkVerified(nil, p2p.Broadcast)

	// Direct send, early exit
	ucp.MarkVerified(b, p2p.Direct)

	// marshal error, early exit
	ucp.MarkVerified(b, p2p.Broadcast)
}
