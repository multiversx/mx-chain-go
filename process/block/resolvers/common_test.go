package resolvers_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
)

func createDataPool() *mock.TransientDataPoolStub {
	transientPool := &mock.TransientDataPoolStub{}
	transientPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transientPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	return transientPool
}

func createRequestMsg(dataType process.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(&process.RequestData{Type: dataType, Value: val})

	return &mock.P2PMessageMock{DataField: buff}
}
