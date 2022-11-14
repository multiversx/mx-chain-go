package requestHandlers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

// HashSliceRequester can request multiple hashes at once
type HashSliceRequester interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ChunkRequester can request a chunk of a large data
type ChunkRequester interface {
	RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error
}

// NonceRequester is a requester that can also request data for a specific nonce and epoch
type NonceRequester interface {
	dataRetriever.Requester
	RequestDataFromNonce(nonce uint64, epoch uint32) error
}

// EpochRequester can request data for a specific epoch
type EpochRequester interface {
	RequestDataFromEpoch(identifier []byte) error
}
