package requestHandlers

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	stubs "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

type recoveryTrieRoutingStub struct {
	stubs.RequesterStub
	calls chan string
}

func (s *recoveryTrieRoutingStub) RequestDataFromHashArray(_ [][]byte, _ uint32) error {
	s.calls <- "normal batch"
	return nil
}

func (s *recoveryTrieRoutingStub) RequestDataFromHashArrayForRecovery(_ [][]byte, _ uint32) error {
	s.calls <- "recovery batch"
	return nil
}

func (s *recoveryTrieRoutingStub) RequestDataFromReferenceAndChunk(_ []byte, _ uint32) error {
	s.calls <- "normal chunk"
	return nil
}

func (s *recoveryTrieRoutingStub) RequestDataFromReferenceAndChunkForRecovery(_ []byte, _ uint32) error {
	s.calls <- "recovery chunk"
	return nil
}

func TestResolverRequestHandler_RecoveryTrieRouting(t *testing.T) {
	requester := &recoveryTrieRoutingStub{calls: make(chan string, 1)}
	handler, err := NewResolverRequestHandler(
		&stubs.RequestersFinderStub{
			MetaChainRequesterCalled:      func(_ string) (dataRetriever.Requester, error) { return requester, nil },
			MetaCrossShardRequesterCalled: func(_ string, _ uint32) (dataRetriever.Requester, error) { return requester, nil },
		}, &mock.RequestedItemsHandlerStub{}, &mock.WhiteListHandlerStub{}, 1000, 0, time.Second, time.Millisecond,
	)
	require.NoError(t, err)
	await := func(expected string) {
		select {
		case actual := <-requester.calls:
			require.Equal(t, expected, actual)
		case <-time.After(time.Second):
			t.Fatal("request was not dispatched")
		}
	}
	for _, recovery := range []bool{false, true, false} {
		handler.SetRecoveryTrieRequests(recovery)
		prefix := "normal"
		if recovery {
			prefix = "recovery"
		}
		handler.lastTrieRequestTime = time.Now().Add(-time.Second)
		handler.RequestTrieNodesForEpoch(0, [][]byte{[]byte("hash")}, "trie", 6617)
		await(prefix + " batch")
		handler.RequestTrieNode([]byte("hash"), "trie", 1)
		await(prefix + " chunk")
	}
}
