package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// HeadersCacheStub -
type HeadersCacheStub struct {
	AddCalled func(round uint64, header *slash.HeaderInfo) error
}

// Add -
func (hcs *HeadersCacheStub) Add(round uint64, header *slash.HeaderInfo) error {
	if hcs.AddCalled != nil {
		return hcs.AddCalled(round, header)
	}
	return nil
}

// IsInterfaceNil -
func (hcs *HeadersCacheStub) IsInterfaceNil() bool {
	return hcs == nil
}
