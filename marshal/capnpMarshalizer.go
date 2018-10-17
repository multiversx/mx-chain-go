package marshal

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/pkg/errors"
	"fmt"
)

type CapnpMarshalizer struct {
}

func (x *CapnpMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	// type assertion on marshalizable structs
	out := bytes.NewBuffer(nil)
	switch t := obj.(type) {
	case *transaction.Transaction:
		o := obj.(*transaction.Transaction)
		// set the transaction members to capnp struct
		o.Save(out)
		break
	case *block.Header:
		o := obj.(*block.Header)
		// set the block members to capnp struct
		o.Save(out)
		break
	case *block.Block:
		o := obj.(*block.Block)
		// set the header members to capnp strunct
		o.Save(out)
		break
	default:
		fmt.Printf("type %v unknown\n", t)
		return nil, errors.New("type not serializable with capnproto")
	}

	return out.Bytes(), nil
}

func (x *CapnpMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	out := bytes.NewBuffer(buff)

	switch obj.(type) {
	case *transaction.Transaction:
		o := obj.(*transaction.Transaction)
		// set the transaction members to capnp struct
		o.Load(out)
		break
	case *block.Header:
		o := obj.(*block.Header)
		// set the block members to capnp struct
		o.Load(out)
		break
	case *block.Block:
		o := obj.(*block.Block)
		// set the header members to capnp strunct
		o.Load(out)
		break
	default:
		return errors.New("type not deserializable with capnproto")
	}

	return nil
}
