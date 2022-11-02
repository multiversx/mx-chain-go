package resolvers_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgRequestingOnlyTrieNodeResolver() resolvers.ArgRequestingOnlyTrieNodeResolver {
	return resolvers.ArgRequestingOnlyTrieNodeResolver{
		ArgBaseResolver: createMockArgBaseResolver(),
	}
}

func TestNewRequestingOnlyTrieNodeResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil topic resolver sender should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgRequestingOnlyTrieNodeResolver()
		arg.SenderResolver = nil
		res, err := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgRequestingOnlyTrieNodeResolver()
		arg.Marshaller = nil
		res, err := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		res, err := resolvers.NewRequestingOnlyTrieNodeResolver(createMockArgRequestingOnlyTrieNodeResolver())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(res))
		assert.Nil(t, res.ProcessReceivedMessage(nil, ""))
	})
}

func TestRequestingOnlyTrieNodeResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	requested := &dataRetriever.RequestData{}
	buffRequested := []byte("node1")

	arg := createMockArgRequestingOnlyTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			assert.Equal(t, buffRequested, rd.Value)
			requested = rd
			return nil
		},
	}
	res, _ := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
	assert.Nil(t, res.RequestDataFromHash(buffRequested, 0))
	assert.Equal(t, &dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: buffRequested,
	}, requested)
}

func TestRequestingOnlyTrieNodeResolver_RequestDataFromHashArray(t *testing.T) {
	t.Parallel()

	t.Run("marshal returns error", func(t *testing.T) {
		t.Parallel()

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		sendRequestCalled := false
		arg := createMockArgRequestingOnlyTrieNodeResolver()
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
				sendRequestCalled = true
				return nil
			},
		}
		arg.Marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		res, _ := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
		err := res.RequestDataFromHashArray([][]byte{hash1, hash2}, 0)
		require.Equal(t, expectedErr, err)
		assert.False(t, sendRequestCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		sendRequestCalled := false
		arg := createMockArgRequestingOnlyTrieNodeResolver()
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
				sendRequestCalled = true
				assert.Equal(t, dataRetriever.HashArrayType, rd.Type)

				b := &batch.Batch{}
				err := arg.Marshaller.Unmarshal(b, rd.Value)
				require.Nil(t, err)
				assert.Equal(t, [][]byte{hash1, hash2}, b.Data)
				assert.Equal(t, uint32(0), b.ChunkIndex) // mandatory to be 0

				return nil
			},
		}
		res, _ := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
		err := res.RequestDataFromHashArray([][]byte{hash1, hash2}, 0)
		require.Nil(t, err)
		assert.True(t, sendRequestCalled)
	})
}

func TestRequestingOnlyTrieNodeResolver_RequestDataFromReferenceAndChunkShouldWork(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	chunkIndex := uint32(343)
	sendRequestCalled := false
	arg := createMockArgRequestingOnlyTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			sendRequestCalled = true
			assert.Equal(t, dataRetriever.HashType, rd.Type)
			assert.Equal(t, hash, rd.Value)
			assert.Equal(t, chunkIndex, rd.ChunkIndex)

			return nil
		},
	}
	res, _ := resolvers.NewRequestingOnlyTrieNodeResolver(arg)
	err := res.RequestDataFromReferenceAndChunk(hash, chunkIndex)
	require.Nil(t, err)
	assert.True(t, sendRequestCalled)
}
