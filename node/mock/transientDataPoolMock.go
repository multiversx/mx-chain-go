package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type TransientDataPoolMock struct {
	HeadersCalled           func() data.ShardedDataCacherNotifier
	HeadersNoncesCalled     func() data.Uint64Cacher
	PeerChangesBlocksCalled func() storage.Cacher
	StateBlocksCalled       func() storage.Cacher
	TransactionsCalled      func() data.ShardedDataCacherNotifier
	TxBlocksCalled          func() storage.Cacher
}

func (tdpm *TransientDataPoolMock) Headers() data.ShardedDataCacherNotifier {
	return tdpm.HeadersCalled()
}

func (tdpm *TransientDataPoolMock) HeadersNonces() data.Uint64Cacher {
	return tdpm.HeadersNoncesCalled()
}

func (tdpm *TransientDataPoolMock) PeerChangesBlocks() storage.Cacher {
	return tdpm.PeerChangesBlocksCalled()
}

func (tdpm *TransientDataPoolMock) StateBlocks() storage.Cacher {
	return tdpm.StateBlocksCalled()
}

func (tdpm *TransientDataPoolMock) Transactions() data.ShardedDataCacherNotifier {
	return tdpm.TransactionsCalled()
}

func (tdpm *TransientDataPoolMock) TxBlocks() storage.Cacher {
	return tdpm.TxBlocksCalled()
}
