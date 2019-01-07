package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type TransientDataPoolStub struct {
	HeadersCalled           func() data.ShardedDataCacherNotifier
	HeadersNoncesCalled     func() data.Uint64Cacher
	PeerChangesBlocksCalled func() storage.Cacher
	StateBlocksCalled       func() storage.Cacher
	TransactionsCalled      func() data.ShardedDataCacherNotifier
	TxBlocksCalled          func() storage.Cacher
}

func (tdps *TransientDataPoolStub) Headers() data.ShardedDataCacherNotifier {
	return tdps.HeadersCalled()
}

func (tdps *TransientDataPoolStub) HeadersNonces() data.Uint64Cacher {
	return tdps.HeadersNoncesCalled()
}

func (tdps *TransientDataPoolStub) PeerChangesBlocks() storage.Cacher {
	return tdps.PeerChangesBlocksCalled()
}

func (tdps *TransientDataPoolStub) StateBlocks() storage.Cacher {
	return tdps.StateBlocksCalled()
}

func (tdps *TransientDataPoolStub) Transactions() data.ShardedDataCacherNotifier {
	return tdps.TransactionsCalled()
}

func (tdps *TransientDataPoolStub) TxBlocks() storage.Cacher {
	return tdps.TxBlocksCalled()
}
