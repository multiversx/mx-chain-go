package mock

import (
	"encoding/json"
	"errors"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/golang/protobuf/proto"
)

// TestingMarshalizers -
var TestingMarshalizers = map[string]marshal.Marshalizer{
	"json":  &JsonMarshalizer{},
	"proto": &ProtobufMarshalizer{},
}

// JsonMarshalizer -
type JsonMarshalizer struct{}

// Marshal -
func (j JsonMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("nil object to serialize from")
	}

	return json.Marshal(obj)
}

// Unmarshal -
func (j JsonMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if obj == nil {
		return errors.New("nil object to serialize to")
	}
	if buff == nil {
		return errors.New("nil byte buffer to deserialize from")
	}
	if len(buff) == 0 {
		return errors.New("empty byte buffer to deserialize from")
	}

	return json.Unmarshal(buff, obj)
}

// IsInterfaceNil -
func (j *JsonMarshalizer) IsInterfaceNil() bool {
	return j == nil
}

// ProtobufMarshalizer -
type ProtobufMarshalizer struct{}

// Marshal -
func (x *ProtobufMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(proto.Message); ok {
		enc, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		return enc, nil
	}
	return nil, errors.New("can not serialize the object")
}

// Unmarshal -
func (x *ProtobufMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(proto.Message); ok {
		return proto.Unmarshal(buff, msg)
	}
	return errors.New("obj does not implement proto.Message")
}

// IsInterfaceNil -
func (x *ProtobufMarshalizer) IsInterfaceNil() bool {
	return x == nil
}

// MarshalizerStub -
type MarshalizerStub struct {
	MarshalCalled   func(obj interface{}) ([]byte, error)
	UnmarshalCalled func(obj interface{}, buff []byte) error
}

// Marshal -
func (ms *MarshalizerStub) Marshal(obj interface{}) ([]byte, error) {
	return ms.MarshalCalled(obj)
}

// Unmarshal -
func (ms *MarshalizerStub) Unmarshal(obj interface{}, buff []byte) error {
	return ms.UnmarshalCalled(obj, buff)
}

// IsInterfaceNil -
func (ms *MarshalizerStub) IsInterfaceNil() bool {
	return ms == nil
}
