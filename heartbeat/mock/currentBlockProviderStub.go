package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// CurrentBlockProviderStub -
type CurrentBlockProviderStub struct {
	GetCurrentBlockHeaderCalled            func() data.HeaderHandler
	SetCurrentBlockHeaderAndRootHashCalled func(bh data.HeaderHandler, rootHash []byte) error
}

// SetCurrentBlockHeaderAndRootHash -
func (cbps *CurrentBlockProviderStub) SetCurrentBlockHeaderAndRootHash(bh data.HeaderHandler, rootHash []byte) error {
	if cbps.SetCurrentBlockHeaderAndRootHashCalled != nil {
		return cbps.SetCurrentBlockHeaderAndRootHashCalled(bh, rootHash)
	}
	return nil
}

// GetCurrentBlockHeader -
func (cbps *CurrentBlockProviderStub) GetCurrentBlockHeader() data.HeaderHandler {
	if cbps.GetCurrentBlockHeaderCalled != nil {
		return cbps.GetCurrentBlockHeaderCalled()
	}
	return nil
}

// IsInterfaceNil -
func (cbps *CurrentBlockProviderStub) IsInterfaceNil() bool {
	return cbps == nil
}
