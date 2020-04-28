package marshal

import (
	"fmt"
)

var _ Marshalizer = (*GogoProtoMarshalizer)(nil)

// GogoProtoMarshalizer implements marshaling with protobuf
type GogoProtoMarshalizer struct {
}

// Marshal does the actual serialization of an object
// The object to be serialized must implement the gogoProtoObj interface
func (x *GogoProtoMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(GogoProtoObj); ok {
		return msg.Marshal()
	}
	return nil, fmt.Errorf("%T, %w", obj, ErrMarshallingProto)
}

// Unmarshal does the actual deserialization of an object
// The object to be deserialized must implement the gogoProtoObj interface
func (x *GogoProtoMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(GogoProtoObj); ok {
		msg.Reset()
		return msg.Unmarshal(buff)
	}

	return fmt.Errorf("%T, %w", obj, ErrUnmarshallingProto)
}

// IsInterfaceNil returns true if there is no value under the interface
func (x *GogoProtoMarshalizer) IsInterfaceNil() bool {
	return x == nil
}
