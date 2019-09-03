package mock

import "github.com/ElrondNetwork/elrond-go/data"

type InterceptedTxHandlerStub struct {
	SndShardCalled    func() uint32
	RcvShardCalled    func() uint32
	HashCalled        func() []byte
	TransactionCalled func() data.TransactionHandler
}

func (itxhs *InterceptedTxHandlerStub) SndShard() uint32 {
	return itxhs.SndShardCalled()
}

func (itxhs *InterceptedTxHandlerStub) RcvShard() uint32 {
	return itxhs.RcvShardCalled()
}

func (itxhs *InterceptedTxHandlerStub) Hash() []byte {
	return itxhs.HashCalled()
}

func (itxhs *InterceptedTxHandlerStub) Transaction() data.TransactionHandler {
	return itxhs.TransactionCalled()
}
