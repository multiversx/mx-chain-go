package mock

// ChunkResolverStub -
type ChunkResolverStub struct {
	HashSliceResolverStub
	RequestDataFromHashAndChunkCalled func(hash []byte, chunkIndex uint32) error
}

// RequestDataFromHashAndChunk -
func (crs *ChunkResolverStub) RequestDataFromHashAndChunk(hash []byte, chunkIndex uint32) error {
	if crs.RequestDataFromHashAndChunkCalled != nil {
		return crs.RequestDataFromHashAndChunkCalled(hash, chunkIndex)
	}

	return nil
}
