package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
)

// NilIndexer will be used when an Indexer is required, but another one isn't necessary or available
type NilIndexer struct {
}

// NewNilIndexer will return a Nil indexer
func NewNilIndexer() *NilIndexer {
	return new(NilIndexer)
}

// SaveBlock will do nothing
func (ni *NilIndexer) SaveBlock(_ data.BodyHandler, _ data.HeaderHandler, _ map[string]data.TransactionHandler, _ []uint64, _ []string) {
}

// SaveMetaBlock will do nothing
func (ni *NilIndexer) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// SaveRoundInfo will do nothing
func (ni *NilIndexer) SaveRoundInfo(_ RoundInfo) {
}

// UpdateTPS will do nothing
func (ni *NilIndexer) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(_ map[uint32][][]byte) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	return ni == nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}
