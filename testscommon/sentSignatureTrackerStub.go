package testscommon

// SentSignatureTrackerStub -
type SentSignatureTrackerStub struct {
	StartRoundCalled                         func()
	SignatureSentCalled                      func(pkBytes []byte)
	ResetCountersForManagedBlockSignerCalled func(signerPk []byte)
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

// ResetCountersForManagedBlockSigner -
func (stub *SentSignatureTrackerStub) ResetCountersForManagedBlockSigner(signerPk []byte) {
	if stub.ResetCountersForManagedBlockSignerCalled != nil {
		stub.ResetCountersForManagedBlockSignerCalled(signerPk)
	}
}

// IsInterfaceNil -
func (stub *SentSignatureTrackerStub) IsInterfaceNil() bool {
	return stub == nil
}
