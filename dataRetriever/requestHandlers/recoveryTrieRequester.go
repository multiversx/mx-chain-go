package requestHandlers

// SetRecoveryTrieRequests includes regular peers while checkpoint tries are restored.
func (rrh *resolverRequestHandler) SetRecoveryTrieRequests(enabled bool) {
	rrh.recoveryTrieRequests.Store(enabled)
}

type recoveryTrieRequester interface {
	RequestDataFromHashArrayForRecovery(hashes [][]byte, epoch uint32) error
	RequestDataFromReferenceAndChunkForRecovery(hash []byte, chunkIndex uint32) error
}

type recoveryTrieRequesterAdapter struct {
	recoveryTrieRequester
}

func (requester *recoveryTrieRequesterAdapter) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.RequestDataFromHashArrayForRecovery(hashes, epoch)
}

func (requester *recoveryTrieRequesterAdapter) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	return requester.RequestDataFromReferenceAndChunkForRecovery(hash, chunkIndex)
}

func (requester *recoveryTrieRequesterAdapter) IsInterfaceNil() bool {
	return requester == nil
}
