package requestHandlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignResolverRequestHandler_ShouldErrNilRequestHandler(t *testing.T) {
	t.Parallel()

	srrh, err := NewSovereignResolverRequestHandler(nil)
	assert.Nil(t, srrh)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewSovereignResolverRequestHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	srrh, err := NewSovereignResolverRequestHandler(rrh)
	assert.NotNil(t, srrh)
	assert.Nil(t, err)
}

func TestSovereignResolverRequestHandler_RequestExtendedShardHeaderByNonce(t *testing.T) {
	requestedNonce := uint64(4)

	shardID := core.SovereignChainShardId
	suffix := fmt.Sprintf("%s_%d", sovUniqueHeadersSuffix, shardID)
	expectedKey := fmt.Sprintf("%d-%d", shardID, requestedNonce)
	expectedFullKey := expectedKey + suffix

	wasWhiteListed := false
	wasNonceRequested := false

	requestItemHandler := &mock.RequestedItemsHandlerStub{
		HasCalled: func(key string) bool {
			require.Equal(t, expectedFullKey, key)
			return false
		},
		AddCalled: func(key string) error {
			require.Equal(t, expectedFullKey, key)
			return nil
		},
	}

	expectedEpoch := uint32(0)
	nonceRequester := &dataRetrieverMocks.NonceRequesterStub{
		RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
			require.Equal(t, requestedNonce, nonce)
			require.Equal(t, expectedEpoch, epoch)

			wasNonceRequested = true
			return nil
		},
	}
	requesterFinder := &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
			require.Equal(t, factory.ExtendedHeaderProofTopic, baseTopic)
			return nonceRequester, nil
		},
	}

	whiteListAdder := &mock.WhiteListHandlerStub{
		AddCalled: func(keys [][]byte) {
			require.Equal(t, keys, [][]byte{[]byte(expectedKey)})
			wasWhiteListed = true
		},
	}

	resolver, _ := NewResolverRequestHandler(
		requesterFinder,
		requestItemHandler,
		whiteListAdder,
		1,
		shardID,
		time.Second,
	)

	sovResolver, _ := NewSovereignResolverRequestHandler(resolver)
	sovResolver.RequestExtendedShardHeaderByNonce(requestedNonce)

	require.True(t, wasNonceRequested)
	require.True(t, wasWhiteListed)
}

func TestSovereignResolverRequestHandler_RequestExtendedShardHeader(t *testing.T) {
	requestedHash := []byte("hash")

	shardID := core.SovereignChainShardId
	suffix := fmt.Sprintf("%s_%d", sovUniqueHeadersSuffix, shardID)
	expectedKey := string(requestedHash)
	expectedFullKey := expectedKey + suffix

	wasWhiteListed := false
	wasHashRequested := false

	requestItemHandler := &mock.RequestedItemsHandlerStub{
		HasCalled: func(key string) bool {
			require.Equal(t, expectedFullKey, key)
			return false
		},
		AddCalled: func(key string) error {
			require.Equal(t, expectedFullKey, key)
			return nil
		},
	}

	expectedEpoch := uint32(0)
	headerRequester := &dataRetrieverMocks.HeaderRequesterStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			require.Equal(t, requestedHash, hash)
			require.Equal(t, expectedEpoch, epoch)

			wasHashRequested = true
			return nil
		},
	}

	requesterFinder := &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
			require.Equal(t, factory.ExtendedHeaderProofTopic, baseTopic)
			return headerRequester, nil
		},
	}

	whiteListAdder := &mock.WhiteListHandlerStub{
		AddCalled: func(keys [][]byte) {
			require.Equal(t, keys, [][]byte{[]byte(expectedKey)})
			wasWhiteListed = true
		},
	}

	resolver, _ := NewResolverRequestHandler(
		requesterFinder,
		requestItemHandler,
		whiteListAdder,
		1,
		shardID,
		time.Second,
	)

	sovResolver, _ := NewSovereignResolverRequestHandler(resolver)
	sovResolver.RequestExtendedShardHeader(requestedHash)

	require.True(t, wasHashRequested)
	require.True(t, wasWhiteListed)
}

func TestSovereignResolverRequestHandler_RequestTrieNode(t *testing.T) {
	requestedHash := []byte("hash")

	expectedTopic := "topic"
	expectedChunk := uint32(4)

	chTxRequested := make(chan struct{})
	wasDataRequested := false
	requesterMock := &dataRetrieverMocks.ChunkRequesterStub{
		RequestDataFromReferenceAndChunkCalled: func(hash []byte, chunkIndex uint32) error {
			require.Equal(t, requestedHash, hash)
			require.Equal(t, expectedChunk, chunkIndex)

			wasDataRequested = true
			chTxRequested <- struct{}{}
			return nil
		},
	}

	requesterFinder := &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
			require.Equal(t, expectedTopic, baseTopic)
			return requesterMock, nil
		},
	}

	resolver, _ := NewResolverRequestHandler(
		requesterFinder,
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		core.SovereignChainShardId,
		time.Second,
	)

	sovResolver, _ := NewSovereignResolverRequestHandler(resolver)
	sovResolver.RequestTrieNode(requestedHash, expectedTopic, expectedChunk)

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestTrieNode")
	}

	require.True(t, wasDataRequested)
}

func TestSovereignResolverRequestHandler_RequestTrieNodes(t *testing.T) {
	requestedHashes := [][]byte{[]byte("hash")}

	expectedTopic := "topic"

	chTxRequested := make(chan struct{})
	wasDataRequested := false
	requesterMock := &dataRetrieverMocks.HashSliceRequesterStub{
		RequestDataFromHashArrayCalled: func(hash [][]byte, epoch uint32) error {
			require.Zero(t, epoch)
			require.Equal(t, requestedHashes, hash)

			wasDataRequested = true
			chTxRequested <- struct{}{}
			return nil
		},
	}

	requesterFinder := &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
			require.Equal(t, expectedTopic, baseTopic)
			return requesterMock, nil
		},
	}

	resolver, _ := NewResolverRequestHandler(
		requesterFinder,
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		core.SovereignChainShardId,
		time.Second,
	)

	sovResolver, _ := NewSovereignResolverRequestHandler(resolver)
	sovResolver.RequestTrieNodes(core.SovereignChainShardId, requestedHashes, expectedTopic)

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestTrieNodes")
	}

	require.True(t, wasDataRequested)
}

func TestSovereignResolverRequestHandler_RequestFromDifferentContainersShouldCallShardBlocksTopic(t *testing.T) {
	t.Parallel()

	expectedTopic := factory.ShardBlocksTopic

	intraShardRequesterCt := 0
	crossShardRequesterCt := 0

	requesterFinder := &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
			intraShardRequesterCt++
			require.Equal(t, expectedTopic, baseTopic)

			switch baseTopic {
			case factory.ShardBlocksTopic:
				return &dataRetrieverMocks.HeaderRequesterStub{}, nil
			case common.ValidatorInfoTopic, factory.MiniBlocksTopic:
				return &dataRetrieverMocks.HashSliceRequesterStub{}, nil
			}

			require.Fail(t, "should not call on other topic")
			return nil, nil
		},
		CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
			crossShardRequesterCt++
			return &dataRetrieverMocks.HeaderRequesterStub{}, nil
		},
	}

	resolver, _ := NewResolverRequestHandler(
		requesterFinder,
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		core.SovereignChainShardId,
		time.Second,
	)
	sovResolver, _ := NewSovereignResolverRequestHandler(resolver)

	sovResolver.RequestStartOfEpochMetaBlock(0)
	require.Equal(t, 1, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestMetaHeader([]byte("hash"))
	require.Equal(t, 2, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestMetaHeaderByNonce(0)
	require.Equal(t, 3, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestShardHeader(core.SovereignChainShardId, []byte("hash"))
	require.Equal(t, 4, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestShardHeaderByNonce(core.SovereignChainShardId, 0)
	require.Equal(t, 5, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	expectedTopic = common.ValidatorInfoTopic
	sovResolver.RequestValidatorInfo([]byte("hash"))
	require.Equal(t, 6, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestValidatorsInfo([][]byte{[]byte("hash")})
	require.Equal(t, 7, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	expectedTopic = factory.MiniBlocksTopic
	sovResolver.RequestMiniBlock(core.SovereignChainShardId, []byte("hash"))
	require.Equal(t, 8, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)

	sovResolver.RequestMiniBlocks(core.SovereignChainShardId, [][]byte{[]byte("hash")})
	require.Equal(t, 9, intraShardRequesterCt)
	require.Zero(t, crossShardRequesterCt)
}
