package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// StatusComponentsMock -
type StatusComponentsMock struct {
	Indexer process.Indexer
}

// ElasticIndexer -
func (scm *StatusComponentsMock) ElasticIndexer() process.Indexer {
	return scm.Indexer
}

// IsInterfaceNil -
func (scm *StatusComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
