package maps

// BucketSortedMapItem defines an item of the bucket sorted map
type BucketSortedMapItem interface {
	GetKey() string
	ComputeScore() uint32
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
}
