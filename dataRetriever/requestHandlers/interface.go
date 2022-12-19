package requestHandlers

// HashSliceRequester can request multiple hashes at once
type HashSliceRequester interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ChunkRequester can request a chunk of a large data
type ChunkRequester interface {
	RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error
}

// NonceRequester can request data for a specific nonce
type NonceRequester interface {
	RequestDataFromNonce(nonce uint64, epoch uint32) error
}

// EpochRequester can request data for a specific epoch
type EpochRequester interface {
	RequestDataFromEpoch(identifier []byte) error
}

// HeaderRequester defines what a block header requester can do
type HeaderRequester interface {
	NonceRequester
	EpochRequester
}
