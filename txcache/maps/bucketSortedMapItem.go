package maps

// BucketSortedMapItem defines an item of the bucket sorted map
type BucketSortedMapItem interface {
	GetKey() string
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
}
