package testscommon

// ExtendedShardHeaderRequestHandlerStub -
type ExtendedShardHeaderRequestHandlerStub struct {
	RequestHandlerStub
	RequestExtendedShardHeaderByNonceCalled func(nonce uint64)
}

// RequestExtendedShardHeaderByNonce -
func (eshrhs *ExtendedShardHeaderRequestHandlerStub) RequestExtendedShardHeaderByNonce(nonce uint64) {
	if eshrhs.RequestExtendedShardHeaderByNonceCalled != nil {
		eshrhs.RequestExtendedShardHeaderByNonceCalled(nonce)
	}
}
