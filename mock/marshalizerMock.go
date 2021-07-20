package mock

import (
	"encoding/json"
	"errors"
)

var errMockMarshalizer = errors.New("MarshalizerMock generic error")

// MarshalizerMock that will be used for testing
type MarshalizerMock struct {
	Fail bool
}

// Marshal converts the input object in a slice of bytes
func (mm *MarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if mm.Fail {
		return nil, errMockMarshalizer
	}

	if obj == nil {
		return nil, errors.New("nil object to serilize from")
	}

	return json.Marshal(obj)
}

// Unmarshal applies the serialized values over an instantiated object
func (mm *MarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if mm.Fail {
		return errMockMarshalizer
	}

	if obj == nil {
		return errors.New("nil object to serilize to")
	}

	if buff == nil {
		return errors.New("nil byte buffer to deserialize from")
	}

	if len(buff) == 0 {
		return errors.New("empty byte buffer to deserialize from")
	}

	return json.Unmarshal(buff, obj)
}

// Version is deprecated and will be removed
func (*MarshalizerMock) Version() string {
	return "JSON/v.0.0.0.1"
}

// IsInterfaceNil returns true if there is no value under the interface
func (mm *MarshalizerMock) IsInterfaceNil() bool {
	return mm == nil
}
