package subRounds

// SubRoundStartExtraSignatureHandlerMock -
type SubRoundStartExtraSignatureHandlerMock struct {
	ResetCalled      func(pubKeys []string) error
	IdentifierCalled func() string
}

// Reset -
func (mock *SubRoundStartExtraSignatureHandlerMock) Reset(pubKeys []string) error {
	if mock.ResetCalled != nil {
		return mock.ResetCalled(pubKeys)
	}
	return nil
}

// Identifier -
func (mock *SubRoundStartExtraSignatureHandlerMock) Identifier() string {
	if mock.IdentifierCalled != nil {
		return mock.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (mock *SubRoundStartExtraSignatureHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
