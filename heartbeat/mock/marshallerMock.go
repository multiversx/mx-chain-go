package mock

import (
	"encoding/json"
	"errors"
)

var errMockMarshaller = errors.New("MarshallerMock generic error")

// MarshallerMock that will be used for testing
type MarshallerMock struct {
	Fail bool
}

// Marshal converts the input object in a slice of bytes
func (mm *MarshallerMock) Marshal(obj interface{}) ([]byte, error) {
	if mm.Fail {
		return nil, errMockMarshaller
	}

	if obj == nil {
		return nil, errors.New("nil object to serilize from")
	}

	return json.Marshal(obj)
}

// Unmarshal applies the serialized values over an instantiated object
func (mm *MarshallerMock) Unmarshal(obj interface{}, buff []byte) error {
	if mm.Fail {
		return errMockMarshaller
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

// IsInterfaceNil returns true if there is no value under the interface
func (mm *MarshallerMock) IsInterfaceNil() bool {
	return mm == nil
}
