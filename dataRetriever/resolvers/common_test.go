package resolvers_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func createDataPool() *mock.PoolsHolderStub {
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	return pools
}

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}
