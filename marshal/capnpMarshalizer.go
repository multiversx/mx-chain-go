package marshal

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// Interface implementing marshaling with capnproto
type CapnpMarshalizer struct {
}

// Marshal does the actual serialization of an object through capnproto
// The object to be serialized must implement the data.CapnpHelper interface
func (x *CapnpMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	out := bytes.NewBuffer(nil)

	o := obj.(data.CapnpHelper)
	// set the members to capnp struct
	o.Save(out)

	return out.Bytes(), nil
}

// Unmarshal does the actual deserialization of an object through capnproto
// The object to be deserialized must implement the data.CapnpHelper interface
func (x *CapnpMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	out := bytes.NewBuffer(buff)

	o := obj.(data.CapnpHelper)
	// set the members to capnp struct
	o.Load(out)

	return nil
}
