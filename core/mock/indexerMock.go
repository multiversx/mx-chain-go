package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body block.Body, header *block.Header)
}

// SaveBlock calls the mock SaveBlock function
func (im *IndexerMock) SaveBlock(body block.Body, header *block.Header) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(body, header)
	}
}
