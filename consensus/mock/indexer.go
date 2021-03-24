package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/indexer"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct{}

// SaveRoundsInfo -
func (im *IndexerMock) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}

// IsInterfaceNil -
func (im *IndexerMock) IsInterfaceNil() bool {
	return false
}
