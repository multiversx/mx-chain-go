package testscommon

import "time"

// RequestHandlerStub -
type RequestHandlerStub struct {
	RequestShardHeaderCalled                    func(shardID uint32, hash []byte)
	RequestShardHeaderForEpochCalled            func(shardID uint32, hash []byte, epoch uint32)
	RequestMetaHeaderCalled                     func(hash []byte)
	RequestMetaHeaderForEpochCalled             func(hash []byte, epoch uint32)
	RequestMetaHeaderByNonceCalled              func(nonce uint64)
	RequestMetaHeaderByNonceForEpochCalled      func(nonce uint64, epoch uint32)
	RequestShardHeaderByNonceCalled             func(shardID uint32, nonce uint64)
	RequestShardHeaderByNonceForEpochCalled     func(shardID uint32, nonce uint64, epoch uint32)
	RequestTransactionHandlerCalled             func(destShardID uint32, txHashes [][]byte)
	RequestTransactionsForEpochCalled           func(destShardID uint32, txHashes [][]byte, epoch uint32)
	RequestScrHandlerCalled                     func(destShardID uint32, txHashes [][]byte)
	RequestScrHandlerForEpochCalled             func(destShardID uint32, txHashes [][]byte, epoch uint32)
	RequestRewardTxHandlerCalled                func(destShardID uint32, txHashes [][]byte)
	RequestRewardTxHandlerForEpochCalled        func(destShardID uint32, rewardTxHashes [][]byte, epoch uint32)
	RequestMiniBlockHandlerCalled               func(destShardID uint32, miniBlockHash []byte)
	RequestMiniBlockForEpochCalled              func(destShardID uint32, miniBlockHash []byte, epoch uint32)
	RequestMiniBlocksHandlerCalled              func(destShardID uint32, miniBlocksHashes [][]byte)
	RequestMiniBlocksForEpochCalled             func(destShardID uint32, miniBlocksHashes [][]byte, epoch uint32)
	RequestTrieNodesCalled                      func(destShardID uint32, hashes [][]byte, topic string)
	RequestStartOfEpochMetaBlockCalled          func(epoch uint32)
	SetNumPeersToQueryCalled                    func(key string, intra int, cross int) error
	GetNumPeersToQueryCalled                    func(key string) (int, int, error)
	RequestTrieNodeCalled                       func(requestHash []byte, topic string, chunkIndex uint32)
	RequestTrieNodesForEpochCalled              func(destShardID uint32, hashes [][]byte, topic string, epoch uint32)
	CreateTrieNodeIdentifierCalled              func(requestHash []byte, chunkIndex uint32) []byte
	RequestPeerAuthenticationsByHashesCalled    func(destShardID uint32, hashes [][]byte)
	RequestValidatorInfoCalled                  func(hash []byte)
	RequestValidatorInfoForEpochCalled          func(hash []byte, epoch uint32)
	RequestValidatorsInfoCalled                 func(hashes [][]byte)
	RequestValidatorsInfoForEpochCalled         func(hashes [][]byte, epoch uint32)
	RequestEquivalentProofByHashCalled          func(headerShard uint32, headerHash []byte)
	RequestEquivalentProofByHashForEpochCalled  func(headerShard uint32, headerHash []byte, epoch uint32)
	RequestEquivalentProofByNonceCalled         func(headerShard uint32, headerNonce uint64)
	RequestEquivalentProofByNonceForEpochCalled func(headerShard uint32, headerNonce uint64, epoch uint32)
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

// SetEpoch -
func (rhs *RequestHandlerStub) SetEpoch(_ uint32) {
}

// RequestShardHeader -
func (rhs *RequestHandlerStub) RequestShardHeader(shardID uint32, hash []byte) {
	if rhs.RequestShardHeaderCalled == nil {
		return
	}
	rhs.RequestShardHeaderCalled(shardID, hash)
}

// RequestShardHeaderForEpoch -
func (rhs *RequestHandlerStub) RequestShardHeaderForEpoch(shardID uint32, hash []byte, epoch uint32) {
	if rhs.RequestShardHeaderForEpochCalled == nil {
		return
	}
	rhs.RequestShardHeaderForEpochCalled(shardID, hash, epoch)
}

// RequestMetaHeader -
func (rhs *RequestHandlerStub) RequestMetaHeader(hash []byte) {
	if rhs.RequestMetaHeaderCalled == nil {
		return
	}
	rhs.RequestMetaHeaderCalled(hash)
}

// RequestMetaHeaderForEpoch -
func (rhs *RequestHandlerStub) RequestMetaHeaderForEpoch(hash []byte, epoch uint32) {
	if rhs.RequestMetaHeaderForEpochCalled == nil {
		return
	}
	rhs.RequestMetaHeaderForEpochCalled(hash, epoch)
}

// RequestMetaHeaderByNonce -
func (rhs *RequestHandlerStub) RequestMetaHeaderByNonce(nonce uint64) {
	if rhs.RequestMetaHeaderByNonceCalled == nil {
		return
	}
	rhs.RequestMetaHeaderByNonceCalled(nonce)
}

// RequestMetaHeaderByNonceForEpoch -
func (rhs *RequestHandlerStub) RequestMetaHeaderByNonceForEpoch(nonce uint64, epoch uint32) {
	if rhs.RequestMetaHeaderByNonceForEpochCalled == nil {
		return
	}
	rhs.RequestMetaHeaderByNonceForEpochCalled(nonce, epoch)
}

// RequestShardHeaderByNonce -
func (rhs *RequestHandlerStub) RequestShardHeaderByNonce(shardID uint32, nonce uint64) {
	if rhs.RequestShardHeaderByNonceCalled == nil {
		return
	}
	rhs.RequestShardHeaderByNonceCalled(shardID, nonce)
}

// RequestShardHeaderByNonceForEpoch -
func (rhs *RequestHandlerStub) RequestShardHeaderByNonceForEpoch(shardID uint32, nonce uint64, epoch uint32) {
	if rhs.RequestShardHeaderByNonceForEpochCalled == nil {
		return
	}
	rhs.RequestShardHeaderByNonceForEpochCalled(shardID, nonce, epoch)
}

// RequestTransactions -
func (rhs *RequestHandlerStub) RequestTransactions(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestTransactionHandlerCalled == nil {
		return
	}
	rhs.RequestTransactionHandlerCalled(destShardID, txHashes)
}

// RequestTransactionsForEpoch -
func (rhs *RequestHandlerStub) RequestTransactionsForEpoch(destShardID uint32, txHashes [][]byte, epoch uint32) {
	if rhs.RequestTransactionsForEpochCalled == nil {
		return
	}
	rhs.RequestTransactionsForEpochCalled(destShardID, txHashes, epoch)
}

// RequestUnsignedTransactions -
func (rhs *RequestHandlerStub) RequestUnsignedTransactions(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestScrHandlerCalled == nil {
		return
	}
	rhs.RequestScrHandlerCalled(destShardID, txHashes)
}

// RequestUnsignedTransactionsForEpoch -
func (rhs *RequestHandlerStub) RequestUnsignedTransactionsForEpoch(destShardID uint32, txHashes [][]byte, epoch uint32) {
	if rhs.RequestScrHandlerForEpochCalled == nil {
		return
	}
	rhs.RequestScrHandlerForEpochCalled(destShardID, txHashes, epoch)
}

// RequestRewardTransactions -
func (rhs *RequestHandlerStub) RequestRewardTransactions(destShardID uint32, txHashes [][]byte) {
	if rhs.RequestRewardTxHandlerCalled == nil {
		return
	}
	rhs.RequestRewardTxHandlerCalled(destShardID, txHashes)
}

func (rhs *RequestHandlerStub) RequestRewardTransactionsForEpoch(destShardID uint32, rewardTxHashes [][]byte, epoch uint32) {
	if rhs.RequestRewardTxHandlerForEpochCalled == nil {
		return
	}
	rhs.RequestRewardTxHandlerForEpochCalled(destShardID, rewardTxHashes, epoch)
}

// RequestMiniBlock -
func (rhs *RequestHandlerStub) RequestMiniBlock(destShardID uint32, miniblockHash []byte) {
	if rhs.RequestMiniBlockHandlerCalled == nil {
		return
	}
	rhs.RequestMiniBlockHandlerCalled(destShardID, miniblockHash)
}

// RequestMiniBlockForEpoch -
func (rhs *RequestHandlerStub) RequestMiniBlockForEpoch(destShardID uint32, miniblockHash []byte, epoch uint32) {
	if rhs.RequestMiniBlockForEpochCalled == nil {
		return
	}
	rhs.RequestMiniBlockForEpochCalled(destShardID, miniblockHash, epoch)
}

// RequestMiniBlocks -
func (rhs *RequestHandlerStub) RequestMiniBlocks(destShardID uint32, miniblocksHashes [][]byte) {
	if rhs.RequestMiniBlocksHandlerCalled == nil {
		return
	}
	rhs.RequestMiniBlocksHandlerCalled(destShardID, miniblocksHashes)
}

// RequestMiniBlocksForEpoch -
func (rhs *RequestHandlerStub) RequestMiniBlocksForEpoch(destShardID uint32, miniblocksHashes [][]byte, epoch uint32) {
	if rhs.RequestMiniBlocksForEpochCalled == nil {
		return
	}
	rhs.RequestMiniBlocksForEpochCalled(destShardID, miniblocksHashes, epoch)
}

// RequestTrieNodes -
func (rhs *RequestHandlerStub) RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string) {
	if rhs.RequestTrieNodesCalled == nil {
		return
	}
	rhs.RequestTrieNodesCalled(destShardID, hashes, topic)
}

// RequestTrieNodesForEpoch -
func (rhs *RequestHandlerStub) RequestTrieNodesForEpoch(destShardID uint32, hashes [][]byte, topic string, epoch uint32) {
	if rhs.RequestTrieNodesForEpochCalled == nil {
		return
	}
	rhs.RequestTrieNodesForEpochCalled(destShardID, hashes, topic, epoch)
}

// CreateTrieNodeIdentifier -
func (rhs *RequestHandlerStub) CreateTrieNodeIdentifier(requestHash []byte, chunkIndex uint32) []byte {
	if rhs.CreateTrieNodeIdentifierCalled != nil {
		return rhs.CreateTrieNodeIdentifierCalled(requestHash, chunkIndex)
	}

	return nil
}

// RequestTrieNode -
func (rhs *RequestHandlerStub) RequestTrieNode(requestHash []byte, topic string, chunkIndex uint32) {
	if rhs.RequestTrieNodeCalled != nil {
		rhs.RequestTrieNodeCalled(requestHash, topic, chunkIndex)
	}
}

// RequestPeerAuthenticationsByHashes -
func (rhs *RequestHandlerStub) RequestPeerAuthenticationsByHashes(destShardID uint32, hashes [][]byte) {
	if rhs.RequestPeerAuthenticationsByHashesCalled != nil {
		rhs.RequestPeerAuthenticationsByHashesCalled(destShardID, hashes)
	}
}

// RequestValidatorInfo -
func (rhs *RequestHandlerStub) RequestValidatorInfo(hash []byte) {
	if rhs.RequestValidatorInfoCalled != nil {
		rhs.RequestValidatorInfoCalled(hash)
	}
}

// RequestValidatorInfoForEpoch -
func (rhs *RequestHandlerStub) RequestValidatorInfoForEpoch(hash []byte, epoch uint32) {
	if rhs.RequestValidatorInfoForEpochCalled != nil {
		rhs.RequestValidatorInfoForEpochCalled(hash, epoch)
	}
}

// RequestValidatorsInfo -
func (rhs *RequestHandlerStub) RequestValidatorsInfo(hashes [][]byte) {
	if rhs.RequestValidatorsInfoCalled != nil {
		rhs.RequestValidatorsInfoCalled(hashes)
	}
}

// RequestValidatorsInfoForEpoch -
func (rhs *RequestHandlerStub) RequestValidatorsInfoForEpoch(hashes [][]byte, epoch uint32) {
	if rhs.RequestValidatorsInfoForEpochCalled != nil {
		rhs.RequestValidatorsInfoForEpochCalled(hashes, epoch)
	}
}

// RequestEquivalentProofByHash -
func (rhs *RequestHandlerStub) RequestEquivalentProofByHash(headerShard uint32, headerHash []byte) {
	if rhs.RequestEquivalentProofByHashCalled != nil {
		rhs.RequestEquivalentProofByHashCalled(headerShard, headerHash)
	}
}

// RequestEquivalentProofByHashForEpoch -
func (rhs *RequestHandlerStub) RequestEquivalentProofByHashForEpoch(headerShard uint32, headerHash []byte, epoch uint32) {
	if rhs.RequestEquivalentProofByHashForEpochCalled != nil {
		rhs.RequestEquivalentProofByHashForEpochCalled(headerShard, headerHash, epoch)
	}
}

// RequestEquivalentProofByNonce -
func (rhs *RequestHandlerStub) RequestEquivalentProofByNonce(headerShard uint32, headerNonce uint64) {
	if rhs.RequestEquivalentProofByNonceCalled != nil {
		rhs.RequestEquivalentProofByNonceCalled(headerShard, headerNonce)
	}
}

// RequestEquivalentProofByNonceForEpoch -
func (rhs *RequestHandlerStub) RequestEquivalentProofByNonceForEpoch(headerShard uint32, headerNonce uint64, epoch uint32) {
	if rhs.RequestEquivalentProofByNonceForEpochCalled != nil {
		rhs.RequestEquivalentProofByNonceForEpochCalled(headerShard, headerNonce, epoch)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhs *RequestHandlerStub) IsInterfaceNil() bool {
	return rhs == nil
}
