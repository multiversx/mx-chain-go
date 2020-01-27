package marshal

import (
	"github.com/gogo/protobuf/proto"
)

// ProtobufMarshalizer implements marshaling with protobuf
type ProtobufMarshalizer struct {
}

// Marshal does the actual serialization of an object through protobuf
func (x *ProtobufMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(proto.Message); ok {
		enc, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		return enc, nil
	}
	return nil, ErrMarshallingProto
}

// Unmarshal does the actual deserialization of an object through protobuf
func (x *ProtobufMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(proto.Message); ok {
		return proto.Unmarshal(buff, msg)
	}
	return ErrUnmarshallingProto
}

// IsInterfaceNil returns true if there is no value under the interface
func (pm *ProtobufMarshalizer) IsInterfaceNil() bool {
	return pm == nil
}
