package mock

type RequestHandlerMock struct {
	RequestTransactionHandlerCalled   func(destShardID uint32, txHashes [][]byte)
	RequestScrHandlerCalled           func(destShardID uint32, txHashes [][]byte)
	RequestMiniBlockHandlerCalled     func(destShardID uint32, miniblockHash []byte)
	RequestHeaderHandlerCalled        func(destShardID uint32, hash []byte)
	RequestHeaderHandlerByNonceCalled func(destShardID uint32, nonce uint64)
}

func (rrh *RequestHandlerMock) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestTransactionHandlerCalled == nil {
		return
	}
	rrh.RequestTransactionHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerMock) RequestUnsignedTransactions(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestScrHandlerCalled == nil {
		return
	}
	rrh.RequestScrHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerMock) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	if rrh.RequestMiniBlockHandlerCalled == nil {
		return
	}
	rrh.RequestMiniBlockHandlerCalled(shardId, miniblockHash)
}

func (rrh *RequestHandlerMock) RequestHeader(shardId uint32, hash []byte) {
	if rrh.RequestHeaderHandlerCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerCalled(shardId, hash)
}

func (rrh *RequestHandlerMock) RequestHeaderByNonce(destShardID uint32, nonce uint64) {
	if rrh.RequestHeaderHandlerByNonceCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerByNonceCalled(destShardID, nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *RequestHandlerMock) IsInterfaceNil() bool {
	if rrh == nil {
		return true
	}
	return false
}
