package mock

import (
	"encoding/json"

	"github.com/pkg/errors"
)

var errMarshalizerFails = errors.New("marshalizerMock generic error")

// MarshalizerMock is used for testing
type MarshalizerMock struct {
	Fail bool
}

// Marshal encodes an object to its byte array representation
func (m *MarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if m.Fail {
		return nil, errMarshalizerFails
	}

	if obj == nil {
		return nil, errors.New("NIL object to serilize from!")
	}

	return json.Marshal(obj)
}

// Unmarshal decodes a byte array and applies the data on an instantiated struct
func (m *MarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
	if m.Fail {
		return errMarshalizerFails
	}

	if obj == nil {
		return errors.New("NIL object to serilize to!")
	}

	if buff == nil {
		return errors.New("NIL byte buffer to deserialize from!")
	}

	if len(buff) == 0 {
		return errors.New("Empty byte buffer to deserialize from!")
	}

	return json.Unmarshal(buff, obj)
}
