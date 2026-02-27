package smartContract

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// ExtractBlockHeaderAndRootHash -
func (service *SCQueryService) ExtractBlockHeaderAndRootHash(query *process.SCQuery) (data.HeaderHandler, []byte, []byte, error) {
	return service.extractBlockHeaderAndRootHash(query)
}
