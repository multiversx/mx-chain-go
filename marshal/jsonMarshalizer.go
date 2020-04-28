package marshal

import (
	"encoding/json"
	"errors"
)

var _ Marshalizer = (*JsonMarshalizer)(nil)

// JsonMarshalizer implements Marshalizer interface using JSON format
type JsonMarshalizer struct {
}

// Marshal tries to serialize obj parameter
func (jm JsonMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("nil object to serialize from")
	}

	return json.Marshal(obj)
}

// Unmarshal tries to deserialize input buffer values into input object
func (jm JsonMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
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

// IsInterfaceNil returns true if there is no value under the interface
func (jm *JsonMarshalizer) IsInterfaceNil() bool {
	return jm == nil
}
