package disabled

import "time"

// RequestHandler implements the RequestHandler interface but does nothing as it is disabled
type RequestHandler struct {
}

// SetEpoch -
func (r *RequestHandler) SetEpoch(_ uint32) {
}

// RequestShardHeader -
func (r *RequestHandler) RequestShardHeader(_ uint32, _ []byte) {
}

// RequestMetaHeader -
func (r *RequestHandler) RequestMetaHeader(_ []byte) {
}

// RequestMetaHeaderByNonce -
func (r *RequestHandler) RequestMetaHeaderByNonce(_ uint64) {
}

// RequestShardHeaderByNonce -
func (r *RequestHandler) RequestShardHeaderByNonce(_ uint32, _ uint64) {
}

// RequestTransaction -
func (r *RequestHandler) RequestTransaction(_ uint32, _ [][]byte) {
}

// RequestUnsignedTransactions -
func (r *RequestHandler) RequestUnsignedTransactions(_ uint32, _ [][]byte) {
}

// RequestRewardTransactions -
func (r *RequestHandler) RequestRewardTransactions(_ uint32, _ [][]byte) {
}

// RequestMiniBlock -
func (r *RequestHandler) RequestMiniBlock(_ uint32, _ []byte) {
}

// RequestMiniBlocks -
func (r *RequestHandler) RequestMiniBlocks(_ uint32, _ [][]byte) {
}

// RequestTrieNodes -
func (r *RequestHandler) RequestTrieNodes(_ uint32, _ [][]byte, _ string) {
}

// RequestStartOfEpochMetaBlock -
func (r *RequestHandler) RequestStartOfEpochMetaBlock(_ uint32) {
}

// RequestInterval -
func (r *RequestHandler) RequestInterval() time.Duration {
	return time.Second
}

// SetNumPeersToQuery -
func (r *RequestHandler) SetNumPeersToQuery(_ string, _ int, _ int) error {
	return nil
}

// GetNumPeersToQuery -
func (r *RequestHandler) GetNumPeersToQuery(_ string) (int, int, error) {
	return 0, 0, nil
}

// IsInterfaceNil -
func (r *RequestHandler) IsInterfaceNil() bool {
	return r == nil
}
