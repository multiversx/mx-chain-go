package mock

import (
	"math/rand"
	"sync"
	"time"
)

// AddressMock implements a mock address generator used in testing
type AddressMock struct {
	bytes []byte
	hash  []byte
}

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewAddressMock generates a new address
func NewAddressMock() *AddressMock {
	buff := make([]byte, HasherMock{}.Size())

	mutex.Lock()
	r.Read(buff)
	mutex.Unlock()

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
