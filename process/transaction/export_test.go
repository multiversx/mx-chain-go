package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func (txi *TxInterceptor) ProcessTx(tx p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	return txi.processTx(tx, rawData)
}

func (inTx *InterceptedTransaction) SetRcvShard(rcvShard uint32) {
	inTx.rcvShard = rcvShard
}

func (inTx *InterceptedTransaction) SetSndShard(sndShard uint32) {
	inTx.sndShard = sndShard
}

func (inTx *InterceptedTransaction) SetIsAddressedToOtherShards(isAddressedToOtherShards bool) {
	inTx.isAddressedToOtherShards = isAddressedToOtherShards
}

func (txRes *TxResolver) ResolveTxRequest(rd process.RequestData) ([]byte, error) {
	return txRes.resolveTxRequest(rd)
}
