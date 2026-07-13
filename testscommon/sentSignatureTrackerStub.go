package testscommon

// SentSignatureTrackerStub -
type SentSignatureTrackerStub struct {
	StartRoundCalled                         func()
	SignatureSentCalled                      func(pkBytes []byte)
	RecordSignedNonceCalled                  func(pkBytes []byte, nonce uint64, headerHash []byte, roundIndex int64)
	GetSignedNonceInfoCalled                 func(pkBytes []byte, nonce uint64) ([]byte, int64, bool)
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
func (stub *SentSignatureTrackerStub) RecordSignedNonce(pkBytes []byte, nonce uint64, headerHash []byte, roundIndex int64) {
	if stub.RecordSignedNonceCalled != nil {
		stub.RecordSignedNonceCalled(pkBytes, nonce, headerHash, roundIndex)
	}
}

// GetSignedNonceInfo -
func (stub *SentSignatureTrackerStub) GetSignedNonceInfo(pkBytes []byte, nonce uint64) ([]byte, int64, bool) {
	if stub.GetSignedNonceInfoCalled != nil {
		return stub.GetSignedNonceInfoCalled(pkBytes, nonce)
	}
	return nil, 0, false
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
