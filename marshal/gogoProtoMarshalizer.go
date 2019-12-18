package marshal

import (
	gproto "github.com/gogo/protobuf/proto"
	proto "github.com/golang/protobuf/proto"
)

type gogoProtoObj interface {
	gproto.Marshaler
	gproto.Unmarshaler
	proto.Message
}

// GogoProtoMarshalizer implements marshaling with protobuf
type GogoProtoMarshalizer struct {
}

// Marshal does the actual serialization of an object through capnproto
// The object to be serialized must implement the data.CapnpHelper interface
func (x *GogoProtoMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(gogoProtoObj); ok {
		return msg.Marshal()
	}
	return nil, ErrMarshallingProto
}

// Unmarshal does the actual deserialization of an object through capnproto
// The object to be deserialized must implement the data.CapnpHelper interface
func (x *GogoProtoMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(gogoProtoObj); ok {
		msg.Reset()
		return msg.Unmarshal(buff)
	}
	return ErrUnmarshallingProto
}

// IsInterfaceNil returns true if there is no value under the interface
func (x *GogoProtoMarshalizer) IsInterfaceNil() bool {
	return x == nil
}
