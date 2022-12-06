package outport

import (
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process"
)

// OutportDataProviderStub -
type OutportDataProviderStub struct {
	PrepareOutportSaveBlockDataCalled func(
		arg process.ArgPrepareOutportSaveBlockData,
	) (*outportcore.ArgsSaveBlockData, error)
}

// PrepareOutportSaveBlockData -
func (a *OutportDataProviderStub) PrepareOutportSaveBlockData(
	arg process.ArgPrepareOutportSaveBlockData,
) (*outportcore.ArgsSaveBlockData, error) {
	if a.PrepareOutportSaveBlockDataCalled != nil {
		return a.PrepareOutportSaveBlockDataCalled(arg)
	}

	return nil, nil
}

// IsInterfaceNil -
func (a *OutportDataProviderStub) IsInterfaceNil() bool {
	return a == nil
}
