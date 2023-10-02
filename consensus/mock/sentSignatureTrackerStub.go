package mock

// SentSignatureTrackerStub -
type SentSignatureTrackerStub struct {
	StartRoundCalled            func()
	SignatureSentCalled         func(pkBytes []byte)
	ReceivedActualSignersCalled func(signersPks []string)
}

// StartRound -
func (stub *SentSignatureTrackerStub) StartRound() {
	if stub.StartRoundCalled != nil {
		stub.StartRoundCalled()
	}
}

// SignatureSent -
func (stub *SentSignatureTrackerStub) SignatureSent(pkBytes []byte) {
	if stub.SignatureSentCalled != nil {
		stub.SignatureSentCalled(pkBytes)
	}
}

// ReceivedActualSigners -
func (stub *SentSignatureTrackerStub) ReceivedActualSigners(signersPks []string) {
	if stub.ReceivedActualSignersCalled != nil {
		stub.ReceivedActualSignersCalled(signersPks)
	}
}

// IsInterfaceNil -
func (stub *SentSignatureTrackerStub) IsInterfaceNil() bool {
	return stub == nil
}
