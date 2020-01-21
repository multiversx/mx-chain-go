package maps

// BucketSortedMapItem is
type BucketSortedMapItem interface {
	GetKey() string
	ComputeScore() uint32
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
}
