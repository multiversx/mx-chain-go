package mock

type RequestHandlerMock struct {
	RequestTransactionHandlerCalled   func(destShardID uint32, txHashes [][]byte)
	RequestMiniBlockHandlerCalled     func(destShardID uint32, miniblockHash []byte)
	RequestHeaderHandlerCalled        func(destShardID uint32, hash []byte)
	RequestHeaderHandlerByNonceCalled func(destShardID uint32, nonce uint64)
}

func (rrh *RequestHandlerMock) RequestTransactionHandler(destShardID uint32, txHashes [][]byte) {
	if rrh.RequestTransactionHandlerCalled == nil {
		return
	}

	rrh.RequestTransactionHandlerCalled(destShardID, txHashes)
}

func (rrh *RequestHandlerMock) RequestMiniBlockHandler(shardId uint32, miniblockHash []byte) {
	if rrh.RequestMiniBlockHandlerCalled == nil {
		return
	}
	rrh.RequestMiniBlockHandlerCalled(shardId, miniblockHash)
}

func (rrh *RequestHandlerMock) RequestHeaderHandler(shardId uint32, hash []byte) {
	if rrh.RequestHeaderHandlerCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerCalled(shardId, hash)
}

func (rrh *RequestHandlerMock) RequestHeaderHandlerByNonce(destShardID uint32, nonce uint64) {
	if rrh.RequestHeaderHandlerByNonceCalled == nil {
		return
	}
	rrh.RequestHeaderHandlerByNonceCalled(destShardID, nonce)
}
