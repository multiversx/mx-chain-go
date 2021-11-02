package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// HeadersCacheStub -
type HeadersCacheStub struct {
	AddCalled func(round uint64, header data.HeaderInfoHandler) error
}

// Add -
func (hcs *HeadersCacheStub) Add(round uint64, header data.HeaderInfoHandler) error {
	if hcs.AddCalled != nil {
		return hcs.AddCalled(round, header)
	}
	return nil
}

// IsInterfaceNil -
func (hcs *HeadersCacheStub) IsInterfaceNil() bool {
	return hcs == nil
}
