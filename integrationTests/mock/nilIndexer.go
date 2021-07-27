package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// NilIndexer will be used when an Indexer is required, but another one isn't necessary or available
type NilIndexer struct {
}

// SaveBlock will do nothing
func (ni *NilIndexer) SaveBlock(_ *indexer.ArgsSaveBlockData) {
}

// SaveValidatorsRating does nothing
func (ni *NilIndexer) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) {
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	return ni == nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}

// RevertIndexedBlock -
func (ni *NilIndexer) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveAccounts -
func (ni *NilIndexer) SaveAccounts(_ uint64, _ []data.UserAccountHandler) {
}

// Close -
func (ni *NilIndexer) Close() error {
	return nil
}

// SaveRoundsInfo -
func (ni *NilIndexer) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}
