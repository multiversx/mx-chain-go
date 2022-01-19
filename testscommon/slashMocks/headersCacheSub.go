package slashMocks

import "github.com/ElrondNetwork/elrond-go-core/data"

// HeadersCacheStub -
type HeadersCacheStub struct {
	AddCalled func(round uint64, hash []byte, header data.HeaderHandler) error
}

// Add -
func (hcs *HeadersCacheStub) Add(round uint64, hash []byte, header data.HeaderHandler) error {
	if hcs.AddCalled != nil {
		return hcs.AddCalled(round, hash, header)
	}
	return nil
}

// IsInterfaceNil -
func (hcs *HeadersCacheStub) IsInterfaceNil() bool {
	return hcs == nil
}
