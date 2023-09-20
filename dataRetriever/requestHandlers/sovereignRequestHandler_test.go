package requestHandlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
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
