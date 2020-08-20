package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// CurrentBlockProviderStub -
type CurrentBlockProviderStub struct {
	GetCurrentBlockHeaderCalled func() data.HeaderHandler
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
