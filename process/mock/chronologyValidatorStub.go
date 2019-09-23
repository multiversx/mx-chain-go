package mock

type ChronologyValidatorStub struct {
	ValidateReceivedBlockCalled func(shardID uint32, epoch uint32, nonce uint64, round uint64) error
}

func (cvs *ChronologyValidatorStub) ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
	return cvs.ValidateReceivedBlockCalled(shardID, epoch, nonce, round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cvs *ChronologyValidatorStub) IsInterfaceNil() bool {
	if cvs == nil {
		return true
	}
	return false
}
