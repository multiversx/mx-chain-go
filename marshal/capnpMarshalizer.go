package marshal

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type CapnpMarshalizer struct {
}

func (x *CapnpMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	// type assertion on marshalizable structs
	out := bytes.NewBuffer(nil)

	o := obj.(data.CapnpHelper)
	// set the transaction members to capnp struct
	o.Save(out)

	return out.Bytes(), nil
}

func (x *CapnpMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	out := bytes.NewBuffer(buff)

	o := obj.(data.CapnpHelper)
	// set the transaction members to capnp struct
	o.Load(out)

	return nil
}
