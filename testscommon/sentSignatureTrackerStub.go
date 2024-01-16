package testscommon

// SentSignatureTrackerStub -
type SentSignatureTrackerStub struct {
	StartRoundCalled                       func()
	SignatureSentCalled                    func(pkBytes []byte)
	ResetCountersManagedBlockSignersCalled func(signersPKs [][]byte)
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

// ResetCountersManagedBlockSigners -
func (stub *SentSignatureTrackerStub) ResetCountersManagedBlockSigners(signersPKs [][]byte) {
	if stub.ResetCountersManagedBlockSignersCalled != nil {
		stub.ResetCountersManagedBlockSignersCalled(signersPKs)
	}
}

// IsInterfaceNil -
func (stub *SentSignatureTrackerStub) IsInterfaceNil() bool {
	return stub == nil
}
