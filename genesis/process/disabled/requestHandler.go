package disabled

import "time"

// RequestHandler implements the RequestHandler interface but does nothing as it is disabled
type RequestHandler struct {
}

// SetEpoch does nothing
func (r *RequestHandler) SetEpoch(_ uint32) {
}

// RequestShardHeader does nothing
func (r *RequestHandler) RequestShardHeader(_ uint32, _ []byte) {
}

// RequestMetaHeader does nothing
func (r *RequestHandler) RequestMetaHeader(_ []byte) {
}

// RequestMetaHeaderByNonce does nothing
func (r *RequestHandler) RequestMetaHeaderByNonce(_ uint64) {
}

// RequestShardHeaderByNonce does nothing
func (r *RequestHandler) RequestShardHeaderByNonce(_ uint32, _ uint64) {
}

// RequestTransaction does nothing
func (r *RequestHandler) RequestTransaction(_ uint32, _ [][]byte) {
}

// RequestUnsignedTransactions does nothing
func (r *RequestHandler) RequestUnsignedTransactions(_ uint32, _ [][]byte) {
}

// RequestRewardTransactions does nothing
func (r *RequestHandler) RequestRewardTransactions(_ uint32, _ [][]byte) {
}

// RequestMiniBlock does nothing
func (r *RequestHandler) RequestMiniBlock(_ uint32, _ []byte) {
}

// RequestMiniBlocks does nothing
func (r *RequestHandler) RequestMiniBlocks(_ uint32, _ [][]byte) {
}

// RequestTrieNodes does nothing
func (r *RequestHandler) RequestTrieNodes(_ uint32, _ [][]byte, _ string) {
}

// RequestStartOfEpochMetaBlock does nothing
func (r *RequestHandler) RequestStartOfEpochMetaBlock(_ uint32) {
}

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
func (r *RequestHandler) RequestTrieNode(_ []byte, _ string, _ uint32) {
}

// CreateTrieNodeIdentifier returns an empty slice
func (r *RequestHandler) CreateTrieNodeIdentifier(_ []byte, _ uint32) []byte {
	return make([]byte, 0)
}

// RequestPeerAuthenticationsByHashes does nothing
func (r *RequestHandler) RequestPeerAuthenticationsByHashes(_ uint32, _ [][]byte) {
}

// RequestValidatorInfo does nothing
func (r *RequestHandler) RequestValidatorInfo(_ []byte) {
}

// RequestValidatorsInfo does nothing
func (r *RequestHandler) RequestValidatorsInfo(_ [][]byte) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *RequestHandler) IsInterfaceNil() bool {
	return r == nil
}
