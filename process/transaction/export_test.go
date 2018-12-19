package transaction

func (inTx *InterceptedTransaction) SetRcvShard(rcvShard uint32) {
	inTx.rcvShard = rcvShard
}

func (inTx *InterceptedTransaction) SetSndShard(sndShard uint32) {
	inTx.sndShard = sndShard
}

func (inTx *InterceptedTransaction) SetIsAddressedToOtherShards(isAddressedToOtherShards bool) {
	inTx.isAddressedToOtherShards = isAddressedToOtherShards
}
