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
func (ni *NilIndexer) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64) {
	return
}

// SaveRoundInfo will do nothing
func (ni *NilIndexer) SaveRoundInfo(round int64, info RoundInfo) {
	return
}

// UpdateTPS will do nothing
func (ni *NilIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	return
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte) {
	return
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	if ni == nil {
		return true
	}
	return false
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}
