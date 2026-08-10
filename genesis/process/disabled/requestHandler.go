package disabled

import "time"

// RequestHandler implements the RequestHandler interface but does nothing as it is disabled
type RequestHandler struct{}

// SetEpoch does nothing
func (r *RequestHandler) SetEpoch(_ uint32) {}

// RequestShardHeader does nothing
func (r *RequestHandler) RequestShardHeader(_ uint32, _ []byte) {}

// RequestShardHeaderForEpoch does nothing
func (r *RequestHandler) RequestShardHeaderForEpoch(_ uint32, _ []byte, _ uint32) {}

// RequestMetaHeader does nothing
func (r *RequestHandler) RequestMetaHeader(_ []byte) {}

// RequestMetaHeaderForEpoch does nothing
func (r *RequestHandler) RequestMetaHeaderForEpoch(hash []byte, epoch uint32) {}

// RequestMetaHeaderByNonce does nothing
func (r *RequestHandler) RequestMetaHeaderByNonce(_ uint64) {}

// RequestMetaHeaderByNonceForEpoch does nothing
func (r *RequestHandler) RequestMetaHeaderByNonceForEpoch(nonce uint64, epoch uint32) {}

// RequestShardHeaderByNonce does nothing
func (r *RequestHandler) RequestShardHeaderByNonce(_ uint32, _ uint64) {}

// RequestShardHeaderByNonceForEpoch does nothing
func (r *RequestHandler) RequestShardHeaderByNonceForEpoch(_ uint32, _ uint64, _ uint32) {}

// RequestTransactions does nothing
func (r *RequestHandler) RequestTransactions(_ uint32, _ [][]byte) {}

// RequestTransactionsForEpoch does nothing
func (r *RequestHandler) RequestTransactionsForEpoch(_ uint32, _ [][]byte, _ uint32) {}

// RequestUnsignedTransactions does nothing
func (r *RequestHandler) RequestUnsignedTransactions(_ uint32, _ [][]byte) {}

// RequestUnsignedTransactionsForEpoch does nothing
func (r *RequestHandler) RequestUnsignedTransactionsForEpoch(_ uint32, _ [][]byte, _ uint32) {}

// RequestRewardTransactions does nothing
func (r *RequestHandler) RequestRewardTransactions(_ uint32, _ [][]byte) {}

// RequestRewardTransactionsForEpoch does nothing
func (r *RequestHandler) RequestRewardTransactionsForEpoch(_ uint32, _ [][]byte, _ uint32) {}

// RequestMiniBlock does nothing
func (r *RequestHandler) RequestMiniBlock(_ uint32, _ []byte) {}

// RequestMiniBlockForEpoch does nothing
func (r *RequestHandler) RequestMiniBlockForEpoch(_ uint32, _ []byte, _ uint32) {}

// RequestMiniBlocks does nothing
func (r *RequestHandler) RequestMiniBlocks(_ uint32, _ [][]byte) {}

// RequestMiniBlocksForEpoch does nothing
func (r *RequestHandler) RequestMiniBlocksForEpoch(_ uint32, _ [][]byte, _ uint32) {}

// RequestTrieNodes does nothing
func (r *RequestHandler) RequestTrieNodes(_ uint32, _ [][]byte, _ string) {}

// RequestTrieNodesForEpoch does nothing
func (r *RequestHandler) RequestTrieNodesForEpoch(_ uint32, _ [][]byte, _ string, _ uint32) {}

// RequestStartOfEpochMetaBlock does nothing
func (r *RequestHandler) RequestStartOfEpochMetaBlock(_ uint32) {}

// RequestInterval returns one second
func (r *RequestHandler) RequestInterval() time.Duration {
	return time.Second
}

// SetNumPeersToQuery returns nil
func (r *RequestHandler) SetNumPeersToQuery(_ string, _ int, _ int) error {
	return nil
}

// GetNumPeersToQuery returns 0, 0 and nil
func (r *RequestHandler) GetNumPeersToQuery(_ string) (int, int, error) {
	return 0, 0, nil
}

// RequestTrieNode does nothing
func (r *RequestHandler) RequestTrieNode(_ []byte, _ string, _ uint32) {}

// CreateTrieNodeIdentifier returns an empty slice
func (r *RequestHandler) CreateTrieNodeIdentifier(_ []byte, _ uint32) []byte {
	return make([]byte, 0)
}

// RequestPeerAuthenticationsByHashes does nothing
func (r *RequestHandler) RequestPeerAuthenticationsByHashes(_ uint32, _ [][]byte) {}

// RequestValidatorInfo does nothing
func (r *RequestHandler) RequestValidatorInfo(_ []byte) {}

// RequestValidatorInfoForEpoch does nothing
func (r *RequestHandler) RequestValidatorInfoForEpoch(_ []byte, _ uint32) {}

// RequestValidatorsInfo does nothing
func (r *RequestHandler) RequestValidatorsInfo(_ [][]byte) {}

// RequestValidatorsInfoForEpoch does nothing
func (r *RequestHandler) RequestValidatorsInfoForEpoch(_ [][]byte, _ uint32) {}

// RequestEquivalentProofByHash does nothing
func (r *RequestHandler) RequestEquivalentProofByHash(_ uint32, _ []byte) {}

// RequestEquivalentProofByHashForEpoch does nothing
func (r *RequestHandler) RequestEquivalentProofByHashForEpoch(_ uint32, _ []byte, _ uint32) {}

// RequestEquivalentProofByNonce does nothing
func (r *RequestHandler) RequestEquivalentProofByNonce(_ uint32, _ uint64) {}

// RequestEquivalentProofByNonceForEpoch does nothing
func (r *RequestHandler) RequestEquivalentProofByNonceForEpoch(_ uint32, _ uint64, _ uint32) {}

// IsInterfaceNil returns true if there is no value under the interface
func (r *RequestHandler) IsInterfaceNil() bool {
	return r == nil
}
