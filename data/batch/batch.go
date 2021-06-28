//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. batch.proto
package batch

// New returns a new batch from given buffers
func New(buffs ...[]byte) *Batch {
	return &Batch{
		Data: buffs,
	}
}

// NewChunk returns a new batch containing a chunk from given buffers
func NewChunk(chunkIndex uint32, reference []byte, maxChunks uint32, buffs ...[]byte) *Batch {
	return &Batch{
		Data:       buffs,
		Reference:  reference,
		ChunkIndex: chunkIndex,
		MaxChunks:  maxChunks,
	}
}
