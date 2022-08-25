package resolvers_test

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}

func createRequestMsgWithChunkIndex(dataType dataRetriever.RequestDataType, val []byte, epoch uint32, chunkIndex uint32) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{
		Type:       dataType,
		Value:      val,
		Epoch:      epoch,
		ChunkIndex: chunkIndex,
	})
	return &mock.P2PMessageMock{DataField: buff}
}
