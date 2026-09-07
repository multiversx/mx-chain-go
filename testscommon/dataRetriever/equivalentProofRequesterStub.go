package dataRetriever

// EquivalentProofRequesterStub -
type EquivalentProofRequesterStub struct {
	*RequesterStub
	RequestDataFromNonceCalled func(nonceShardKey []byte, epoch uint32) error
}

// RequestDataFromNonce -
func (stub *EquivalentProofRequesterStub) RequestDataFromNonce(nonceShardKey []byte, epoch uint32) error {
	if stub.RequestDataFromNonceCalled != nil {
		return stub.RequestDataFromNonceCalled(nonceShardKey, epoch)
	}
	return nil
}
