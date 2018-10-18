package data

import (
	"io"
	"math/rand"
	"fmt"
)

// CapnpHelper is an interface that defines methods needed for
// serializing and deserializing Capnp structures into Go structures and viceversa
type CapnpHelper interface {
	// Save saves the serialized data of the implementer type into a stream through Capnp protocol
	Save(w io.Writer) error
	// Load loads the data from the stream into a go structure through Capnp protocol
	Load(r io.Reader) error
}

// DataGenerator is an interface for defining dummy array of data of each implementer type
// Used for tests
type DataGenerator interface {
	// GenerateDummyArray generates an array of data of the implementer type
	// The implementer needs to implement CapnpHelper as well
	GenerateDummyArray() []CapnpHelper
}

// RandomStr generates random strings of set length
func RandomStr(l int) string {
	buf := make([]byte, l)

	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}
