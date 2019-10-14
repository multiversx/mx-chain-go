package data

import (
	"io"
)

// CapnpHelper is an interface that defines methods needed for
// serializing and deserializing Capnp structures into Go structures and viceversa
type CapnpHelper interface {
	// Save saves the serialized data of the implementer type into a stream through Capnp protocol
	Save(w io.Writer) error
	// Load loads the data from the stream into a go structure through Capnp protocol
	Load(r io.Reader) error
}
