package subRounds

import "github.com/multiversx/mx-chain-go/consensus"

// SubRoundStartExtraSignersHolderMock -
type SubRoundStartExtraSignersHolderMock struct {
	ResetCalled                       func(pubKeys []string) error
	RegisterExtraSigningHandlerCalled func(extraSigner consensus.SubRoundStartExtraSignatureHandler) error
}

// Reset -
func (mock *SubRoundStartExtraSignersHolderMock) Reset(pubKeys []string) error {
	if mock.ResetCalled != nil {
		return mock.ResetCalled(pubKeys)
	}
	return nil
}

// RegisterExtraSigningHandler -
func (mock *SubRoundStartExtraSignersHolderMock) RegisterExtraSigningHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error {
	if mock.RegisterExtraSigningHandlerCalled != nil {
		return mock.RegisterExtraSigningHandlerCalled(extraSigner)
	}
	return nil
}

// IsInterfaceNil -
func (mock *SubRoundStartExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
