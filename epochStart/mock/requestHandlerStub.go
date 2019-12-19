package mock

type RequestHandlerStub struct {
	RequestTransactionHandlerCalled   func(destShardID uint32, txHashes [][]byte)
	RequestScrHandlerCalled           func(destShardID uint32, txHashes [][]byte)
	RequestRewardTxHandlerCalled      func(destShardID uint32, txHashes [][]byte)
	RequestMiniBlockHandlerCalled     func(destShardID uint32, miniblockHash []byte)
	RequestHeaderHandlerCalled        func(destShardID uint32, hash []byte)
	RequestHeaderHandlerByNonceCalled func(destShardID uint32, nonce uint64)
}

func (rrh *RequestHandlerStub) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestTransactionHandlerCalled == nil {
		return
	}
	rrh.RequestTransactionHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerStub) RequestUnsignedTransactions(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestScrHandlerCalled == nil {
		return
	}
	rrh.RequestScrHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerStub) RequestRewardTransactions(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestRewardTxHandlerCalled == nil {
		return
	}
	rrh.RequestRewardTxHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerStub) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	if rrh.RequestMiniBlockHandlerCalled == nil {
		return
	}
	rrh.RequestMiniBlockHandlerCalled(shardId, miniblockHash)
}

func (rrh *RequestHandlerStub) RequestHeader(shardId uint32, hash []byte) {
	if rrh.RequestHeaderHandlerCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerCalled(shardId, hash)
}

func (rrh *RequestHandlerStub) RequestHeaderByNonce(destShardID uint32, nonce uint64) {
	if rrh.RequestHeaderHandlerByNonceCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerByNonceCalled(destShardID, nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *RequestHandlerStub) IsInterfaceNil() bool {
	return rrh == nil
}
