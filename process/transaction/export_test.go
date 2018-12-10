package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (txi *transactionInterceptor) ProcessTx(tx p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	return txi.processTx(tx, rawData, hasher)
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
