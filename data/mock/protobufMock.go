package mock

import (
	"errors"

	"github.com/gogo/protobuf/proto"
)

// ProtobufMarshalizer implements marshaling with protobuf
type ProtobufMarshalizerMock struct {
}

// Marshal does the actual serialization of an object through capnproto
// The object to be serialized must implement the data.CapnpHelper interface
func (x *ProtobufMarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(proto.Message); ok {
		enc, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		return enc, nil
	}
	return nil, errors.New("can not serialize the object")
}

// Unmarshal does the actual deserialization of an object through capnproto
// The object to be deserialized must implement the data.CapnpHelper interface
func (x *ProtobufMarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(proto.Message); ok {
		return proto.Unmarshal(buff, msg)
	}
	return errors.New("obj does not implement proto.Message")
}

// IsInterfaceNil returns true if there is no value under the interface
func (x *ProtobufMarshalizerMock) IsInterfaceNil() bool {
	if x == nil {
		return true
	}
	return false
}
