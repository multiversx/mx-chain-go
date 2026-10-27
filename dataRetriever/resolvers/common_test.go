package resolvers_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func TestDeduplicateHashes(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hash3 := []byte("hash3")

	hashes := [][]byte{
		hash1,
		hash2,
		hash1,
		hash3,
		hash2,
	}

	uniqueHashes := resolvers.DeduplicateHashes(hashes)

	require.Equal(t, [][]byte{
		hash1,
		hash2,
		hash3,
	}, uniqueHashes)
}

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &p2pmocks.P2PMessageMock{DataField: buff}
}

func createRequestMsgWithChunkIndex(dataType dataRetriever.RequestDataType, val []byte, epoch uint32, chunkIndex uint32) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{
		Type:       dataType,
		Value:      val,
		Epoch:      epoch,
		ChunkIndex: chunkIndex,
	})
	return &p2pmocks.P2PMessageMock{DataField: buff}
}
