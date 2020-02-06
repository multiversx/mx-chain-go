package maps

import "github.com/ElrondNetwork/elrond-go/core/atomic"

// BucketSortedMapItem is
type BucketSortedMapItem interface {
	GetKey() string
	ComputeScore() uint32
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
	ScoreChangeInProgressFlag() *atomic.Flag
}
