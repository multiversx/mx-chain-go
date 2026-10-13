package requesters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	stubs "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

type recoverySenderStub struct {
	*stubs.TopicRequestSenderStub
	request *dataRetriever.RequestData
}

func (s *recoverySenderStub) SendOnRequestTopicIncludingMainPeers(rd *dataRetriever.RequestData, _ [][]byte) error {
	s.request = rd
	return nil
}

func TestTrieNodeRequester_RecoveryRouting(t *testing.T) {
	marshalizer := &marshal.GogoProtoMarshalizer{}
	ordinaryCalls := 0
	sender := &recoverySenderStub{TopicRequestSenderStub: &stubs.TopicRequestSenderStub{
		SendOnRequestTopicCalled: func(_ *dataRetriever.RequestData, _ [][]byte) error { ordinaryCalls++; return nil },
	}}
	requester, err := NewTrieNodeRequester(ArgTrieNodeRequester{ArgBaseRequester{RequestSender: sender, Marshaller: marshalizer}})
	require.NoError(t, err)
	hashes := [][]byte{[]byte("hash")}
	require.NoError(t, requester.RequestDataFromHashArrayForRecovery(hashes, 6617))
	require.Equal(t, uint32(6617), sender.request.Epoch)
	decoded := &batch.Batch{}
	require.NoError(t, marshalizer.Unmarshal(decoded, sender.request.Value))
	require.Equal(t, hashes, decoded.Data)
	require.NoError(t, requester.RequestDataFromReferenceAndChunkForRecovery(hashes[0], 3))
	require.Equal(t, dataRetriever.HashType, sender.request.Type)
	require.Equal(t, hashes[0], sender.request.Value)
	require.Equal(t, uint32(3), sender.request.ChunkIndex)
	require.Zero(t, ordinaryCalls)
	require.NoError(t, requester.RequestDataFromHashArray(hashes, 6617))
	require.NoError(t, requester.RequestDataFromReferenceAndChunk(hashes[0], 3))
	require.Equal(t, 2, ordinaryCalls)
}
