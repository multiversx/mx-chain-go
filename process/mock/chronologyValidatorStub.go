package mock

type ChronologyValidatorStub struct {
	ValidateReceivedBlockCalled func(shardID uint32, epoch uint32, nonce uint64, round uint32) error
}

func (cvs *ChronologyValidatorStub) ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
	return cvs.ValidateReceivedBlockCalled(shardID, epoch, nonce, round)
}
