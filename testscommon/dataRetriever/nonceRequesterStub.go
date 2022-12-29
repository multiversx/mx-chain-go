package dataRetriever

// NonceRequesterStub -
type NonceRequesterStub struct {
	RequesterStub
	RequestDataFromNonceCalled func(nonce uint64, epoch uint32) error
}

// RequestDataFromNonce -
func (stub *NonceRequesterStub) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	if stub.RequestDataFromNonceCalled != nil {
		return stub.RequestDataFromNonceCalled(nonce, epoch)
	}
	return nil
}
