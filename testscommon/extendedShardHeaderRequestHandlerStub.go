package testscommon

type ExtendedShardHeaderRequestHandlerStub struct {
	RequestHandlerStub
	RequestExtendedShardHeaderByNonceCalled func(nonce uint64)
}

// RequestShardHeaderByNonce -
func (eshrhs *ExtendedShardHeaderRequestHandlerStub) RequestExtendedShardHeaderByNonce(nonce uint64) {
	if eshrhs.RequestExtendedShardHeaderByNonceCalled != nil {
		eshrhs.RequestExtendedShardHeaderByNonceCalled(nonce)
	}
}
