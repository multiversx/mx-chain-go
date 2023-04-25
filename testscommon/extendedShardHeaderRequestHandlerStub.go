package testscommon

// ExtendedShardHeaderRequestHandlerStub -
type ExtendedShardHeaderRequestHandlerStub struct {
	RequestHandlerStub
	RequestExtendedShardHeaderByNonceCalled func(nonce uint64)
	RequestExtendedShardHeaderCalled        func(hash []byte)
}

// RequestExtendedShardHeaderByNonce -
func (eshrhs *ExtendedShardHeaderRequestHandlerStub) RequestExtendedShardHeaderByNonce(nonce uint64) {
	if eshrhs.RequestExtendedShardHeaderByNonceCalled != nil {
		eshrhs.RequestExtendedShardHeaderByNonceCalled(nonce)
	}
}

// RequestExtendedShardHeader -
func (eshrhs *ExtendedShardHeaderRequestHandlerStub) RequestExtendedShardHeader(hash []byte) {
	if eshrhs.RequestExtendedShardHeaderCalled != nil {
		eshrhs.RequestExtendedShardHeaderCalled(hash)
	}
}
