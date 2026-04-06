package testscommon

// SentSignatureTrackerStub -
type SentSignatureTrackerStub struct {
	StartRoundCalled                         func()
	SignatureSentCalled                      func(pkBytes []byte)
	RecordSignedNonceCalled                  func(pkBytes []byte, nonce uint64, headerHash []byte)
	GetSignedHashCalled                      func(pkBytes []byte, nonce uint64) ([]byte, bool)
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

// RecordSignedNonce -
func (stub *SentSignatureTrackerStub) RecordSignedNonce(pkBytes []byte, nonce uint64, headerHash []byte) {
	if stub.RecordSignedNonceCalled != nil {
		stub.RecordSignedNonceCalled(pkBytes, nonce, headerHash)
	}
}

// GetSignedHash -
func (stub *SentSignatureTrackerStub) GetSignedHash(pkBytes []byte, nonce uint64) ([]byte, bool) {
	if stub.GetSignedHashCalled != nil {
		return stub.GetSignedHashCalled(pkBytes, nonce)
	}
	return nil, false
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
