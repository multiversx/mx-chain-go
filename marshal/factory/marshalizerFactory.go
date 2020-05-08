package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/marshal"
)

// JsonMarshalizer is the name reserved for the json marshalizer
const JsonMarshalizer = "json"

// TxJsonMarshalizer is the name reserved for the transaction json marshalizer
const TxJsonMarshalizer = "tx-json"

// GogoProtobuf is the name reserved for the gogoslick protobuf marshalizer
const GogoProtobuf = "gogo protobuf"

// NewMarshalizer creates a new marshalizer instance based on the provided parameters
func NewMarshalizer(name string) (marshal.Marshalizer, error) {
	switch name {
	case JsonMarshalizer:
		return &marshal.JsonMarshalizer{}, nil
	case GogoProtobuf:
		return &marshal.GogoProtoMarshalizer{}, nil
	case TxJsonMarshalizer:
		return &marshal.TxJsonMarshalizer{}, nil
	default:
		return nil, fmt.Errorf("%w '%s'", marshal.ErrUnknownMarshalizer, name)
	}
}
