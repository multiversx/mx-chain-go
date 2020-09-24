package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
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

// RevertIndexedBlock will do nothing
func (ni *NilIndexer) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SetTxLogsProcessor will do nothing
func (ni *NilIndexer) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// SaveRoundsInfo will do nothing
func (ni *NilIndexer) SaveRoundsInfo(_ []workItems.RoundInfo) {
}

// UpdateTPS will do nothing
func (ni *NilIndexer) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsRating --
func (ni *NilIndexer) SaveValidatorsRating(_ string, _ []workItems.ValidatorRatingInfo) {
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveAccounts won't do anything as this is a nil implementation
func (ni *NilIndexer) SaveAccounts(_ []state.UserAccountHandler) {
}

// Close will do nothing
func (ni *NilIndexer) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	return ni == nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}
