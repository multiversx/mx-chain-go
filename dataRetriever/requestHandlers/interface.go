package requestHandlers

// HashSliceResolver can request multiple hashes at once
type HashSliceResolver interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ChunkResolver can request a chunk of a large data
type ChunkResolver interface {
	RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error
}
