package mock

// ChunkResolverStub -
type ChunkResolverStub struct {
	HashSliceResolverStub
	RequestDataFromReferenceAndChunkCalled func(hash []byte, chunkIndex uint32) error
}

// RequestDataFromReferenceAndChunk -
func (crs *ChunkResolverStub) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	if crs.RequestDataFromReferenceAndChunkCalled != nil {
		return crs.RequestDataFromReferenceAndChunkCalled(hash, chunkIndex)
	}

	return nil
}
