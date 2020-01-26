package marshal

import (
	"bytes"
)

// CapnpMarshalizer implements marshaling with capnproto
type CapnpMarshalizer struct {
}

// Marshal does the actual serialization of an object through capnproto
// The object to be serialized must implement the data.CapnpHelper interface
func (cm *CapnpMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	out := bytes.NewBuffer(nil)

	o := obj.(CapnpHelper)
	// set the members to capnp struct
	err := o.Save(out)

	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// Unmarshal does the actual deserialization of an object through capnproto
// The object to be deserialized must implement the data.CapnpHelper interface
func (cm *CapnpMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	out := bytes.NewBuffer(buff)

	o := obj.(CapnpHelper)
	// set the members to capnp struct
	err := o.Load(out)

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *CapnpMarshalizer) IsInterfaceNil() bool {
	return cm == nil
}
