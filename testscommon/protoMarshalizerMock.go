package testscommon

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/marshal"
)

var _ marshal.Marshalizer = (*ProtoMarshalizerMock)(nil)

// ProtoMarshalizerMock implements marshaling with protobuf
type ProtoMarshalizerMock struct {
}

// Marshal does the actual serialization of an object
// The object to be serialized must implement the gogoProtoObj interface
func (pmm *ProtoMarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if msg, ok := obj.(marshal.GogoProtoObj); ok {
		return msg.Marshal()
	}
	return nil, fmt.Errorf("%T, %w", obj, marshal.ErrMarshallingProto)
}

// Unmarshal does the actual deserialization of an object
// The object to be deserialized must implement the gogoProtoObj interface
func (pmm *ProtoMarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if msg, ok := obj.(marshal.GogoProtoObj); ok {
		msg.Reset()
		return msg.Unmarshal(buff)
	}

	return fmt.Errorf("%T, %w", obj, marshal.ErrUnmarshallingProto)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pmm *ProtoMarshalizerMock) IsInterfaceNil() bool {
	return pmm == nil
}
