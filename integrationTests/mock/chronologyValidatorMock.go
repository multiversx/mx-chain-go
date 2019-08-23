package mock

type ChronologyValidatorMock struct {
}

func (cvm *ChronologyValidatorMock) ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cvm *ChronologyValidatorMock) IsInterfaceNil() bool {
	if cvm == nil {
		return true
	}
	return false
}
