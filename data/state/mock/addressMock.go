package mock

import (
	"math/rand"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// AddressMock implements a mock address generator used in testing
type AddressMock struct {
	bytes []byte
	hash  []byte
}

// NewAddressMock generates a new address
func NewAddressMock() *AddressMock {
	buff := make([]byte, state.AdrLen)
	rand.Read(buff)

	return &AddressMock{bytes: buff}
}

// Bytes returns the address' bytes
func (address *AddressMock) Bytes() []byte {
	return address.bytes
}

// Hash returns the address hash
func (address *AddressMock) Hash() []byte {
	if address.hash == nil {
		address.hash = HasherMock{}.Compute(string(address.bytes))
	}

	return address.hash
}
