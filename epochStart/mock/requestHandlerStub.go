package mock

import "time"

// RequestHandlerStub -
type RequestHandlerStub struct {
	RequestShardHeaderCalled           func(shardId uint32, hash []byte)
	RequestMetaHeaderCalled            func(hash []byte)
	RequestMetaHeaderByNonceCalled     func(nonce uint64)
	RequestShardHeaderByNonceCalled    func(shardId uint32, nonce uint64)
	RequestTransactionHandlerCalled    func(destShardID uint32, txHashes [][]byte)
	RequestScrHandlerCalled            func(destShardID uint32, txHashes [][]byte)
	RequestRewardTxHandlerCalled       func(destShardID uint32, txHashes [][]byte)
	RequestMiniBlocksHandlerCalled     func(destShardID uint32, miniblockHashes [][]byte)
	RequestStartOfEpochMetaBlockCalled func(epoch uint32)
	SetNumPeersToQueryCalled           func(key string, intra int, cross int) error
	GetNumPeersToQueryCalled           func(key string) (int, int, error)
}

// SetEpoch -
func (rhs *RequestHandlerStub) SetEpoch(_ uint32) {
	panic("implement me")
}

// RequestMiniBlock -
func (rhs *RequestHandlerStub) RequestMiniBlock(_ uint32, _ []byte) {
	panic("implement me")
}

// RequestTrieNodes -
func (rhs *RequestHandlerStub) RequestTrieNodes(_ uint32, _ [][]byte, _ string) {
}

// SetNumPeersToQuery -
func (rhs *RequestHandlerStub) SetNumPeersToQuery(key string, intra int, cross int) error {
	if rhs.SetNumPeersToQueryCalled != nil {
		return rhs.SetNumPeersToQueryCalled(key, intra, cross)
	}

	return nil
}

// GetNumPeersToQuery -
func (rhs *RequestHandlerStub) GetNumPeersToQuery(key string) (int, int, error) {
	if rhs.GetNumPeersToQueryCalled != nil {
		return rhs.GetNumPeersToQueryCalled(key)
	}

	return 2, 2, nil
}

// RequestInterval -
func (rhs *RequestHandlerStub) RequestInterval() time.Duration {
	return time.Second
}

// RequestStartOfEpochMetaBlock -
func (rhs *RequestHandlerStub) RequestStartOfEpochMetaBlock(epoch uint32) {
	if rhs.RequestStartOfEpochMetaBlockCalled == nil {
		return
	}
	rhs.RequestStartOfEpochMetaBlockCalled(epoch)
}

// RequestShardHeader -
func (rhs *RequestHandlerStub) RequestShardHeader(shardId uint32, hash []byte) {
	if rhs.RequestShardHeaderCalled == nil {
		return
	}
	rhs.RequestShardHeaderCalled(shardId, hash)
}

// RequestMetaHeader -
func (rhs *RequestHandlerStub) RequestMetaHeader(hash []byte) {
	if rhs.RequestMetaHeaderCalled == nil {
		return
	}
	rhs.RequestMetaHeaderCalled(hash)
}

// RequestMetaHeaderByNonce -
func (rhs *RequestHandlerStub) RequestMetaHeaderByNonce(nonce uint64) {
	if rhs.RequestMetaHeaderByNonceCalled == nil {
		return
	}
	rhs.RequestMetaHeaderByNonceCalled(nonce)
}

// RequestShardHeaderByNonce -
func (rhs *RequestHandlerStub) RequestShardHeaderByNonce(shardId uint32, nonce uint64) {
	if rhs.RequestShardHeaderByNonceCalled == nil {
		return
	}
	rhs.RequestShardHeaderByNonceCalled(shardId, nonce)
}

// RequestTransaction -
func (rhs *RequestHandlerStub) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestTransactionHandlerCalled == nil {
		return
	}
	rhs.RequestTransactionHandlerCalled(destShardID, txHashes)
}

// RequestUnsignedTransactions -
func (rhs *RequestHandlerStub) RequestUnsignedTransactions(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestScrHandlerCalled == nil {
		return
	}
	rhs.RequestScrHandlerCalled(destShardID, txHashes)
}

// RequestRewardTransactions -
func (rhs *RequestHandlerStub) RequestRewardTransactions(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestRewardTxHandlerCalled == nil {
		return
	}
	rhs.RequestRewardTxHandlerCalled(destShardID, txHashes)
}

// RequestMiniBlock -
func (rhs *RequestHandlerStub) RequestMiniBlocks(shardId uint32, miniblockHashes [][]byte) {
	if rhs.RequestMiniBlocksHandlerCalled == nil {
		return
	}
	rhs.RequestMiniBlocksHandlerCalled(shardId, miniblockHashes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhs *RequestHandlerStub) IsInterfaceNil() bool {
	return rhs == nil
}
