package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// OutportDataProviderStub -
type OutportDataProviderStub struct {
	PrepareOutportSaveBlockDataCalled func(
		headerHash []byte,
		body data.BodyHandler,
		header data.HeaderHandler,
		rewardsTxs map[string]data.TransactionHandler,
		notarizedHeadersHashes []string,
	) (*indexer.ArgsSaveBlockData, error)
}

// PrepareOutportSaveBlockData -
func (a *OutportDataProviderStub) PrepareOutportSaveBlockData(
	headerHash []byte,
	body data.BodyHandler,
	header data.HeaderHandler,
	rewardsTxs map[string]data.TransactionHandler,
	notarizedHeadersHashes []string,
) (*indexer.ArgsSaveBlockData, error) {
	if a.PrepareOutportSaveBlockDataCalled != nil {
		return a.PrepareOutportSaveBlockDataCalled(headerHash, body, header, rewardsTxs, notarizedHeadersHashes)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *OutportDataProviderStub) IsInterfaceNil() bool {
	return a == nil
}
