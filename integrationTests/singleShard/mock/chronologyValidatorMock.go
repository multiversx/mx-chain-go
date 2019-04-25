package mock

type ChronologyValidatorMock struct {
}

func (cvm *ChronologyValidatorMock) ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
	return nil
}
