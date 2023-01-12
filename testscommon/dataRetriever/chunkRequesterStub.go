package dataRetriever

// ChunkRequesterStub -
type ChunkRequesterStub struct {
	RequesterStub
	RequestDataFromReferenceAndChunkCalled func(reference []byte, chunkIndex uint32) error
}

// RequestDataFromReferenceAndChunk -
func (stub *ChunkRequesterStub) RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error {
	if stub.RequestDataFromReferenceAndChunkCalled != nil {
		return stub.RequestDataFromReferenceAndChunkCalled(reference, chunkIndex)
	}
	return nil
}
