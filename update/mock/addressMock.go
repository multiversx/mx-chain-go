package mock

import (
	"math/rand"
	"sync"
	"time"
)

// AddressMock implements a mock address generator used in testing
type AddressMock struct {
	bytes []byte
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
	_, _ = r.Read(buff)
	mutex.Unlock()

	return &AddressMock{bytes: buff}
}

// Bytes returns the address' bytes
func (address *AddressMock) Bytes() []byte {
	return address.bytes
}

// IsInterfaceNil returns true if there is no value under the interface
func (address *AddressMock) IsInterfaceNil() bool {
	return address == nil
}
