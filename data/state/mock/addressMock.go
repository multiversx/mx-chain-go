package mock

import (
	"math/rand"
	"time"
)

// AddressMock implements a mock address generator used in testing
type AddressMock struct {
	bytes []byte
	hash  []byte
}

var r *rand.Rand

// NewAddressMock generates a new address
func NewAddressMock() *AddressMock {
	buff := make([]byte, HasherMock{}.Size())

	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	r.Read(buff)

	return &AddressMock{bytes: buff}
}

// NewAddressMockFromBytes generates a new address
func NewAddressMockFromBytes(buff []byte) *AddressMock {
	return &AddressMock{bytes: buff}
}

// Bytes returns the address' bytes
func (address *AddressMock) Bytes() []byte {
	return address.bytes
}
