package mock

import "github.com/ElrondNetwork/elrond-go-core/data"

// HeadersCacheStub -
type HeadersCacheStub struct {
	ContainsCalled func(round uint64, hash []byte) bool
}

// Add -
func (hcs *HeadersCacheStub) Add(round uint64, hash []byte, header data.HeaderHandler) {
}

// Contains -
func (hcs *HeadersCacheStub) Contains(round uint64, hash []byte) bool {
	if hcs.ContainsCalled != nil {
		return hcs.ContainsCalled(round, hash)
	}
	return false
}

// IsInterfaceNil -
func (hcs *HeadersCacheStub) IsInterfaceNil() bool {
	return hcs == nil
}
