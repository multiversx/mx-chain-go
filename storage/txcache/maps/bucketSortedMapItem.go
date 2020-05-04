package maps

import (
	"sync/atomic"
)

// BucketSortedMapItem defines an item of the bucket sorted map
type BucketSortedMapItem interface {
	GetKey() string
	ComputeScore() uint32
	ScoreChunk() *MapChunkPointer
}

// MapChunkPointer is an atomically-mutable pointer to a MapChunk
type MapChunkPointer struct {
	value atomic.Value
}

// Set sets the value
func (variable *MapChunkPointer) Set(value *MapChunk) {
	variable.value.Store(value)
}

// Get gets the value
func (variable *MapChunkPointer) Get() *MapChunk {
	value := variable.value.Load()
	if value == nil {
		return nil
	}

	return value.(*MapChunk)
}
