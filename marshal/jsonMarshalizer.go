package marshal

import (
	"encoding/json"
	"errors"
)

// JsonMarshalizer implements Marshalizer interface using JSON format
type JsonMarshalizer struct {
}

// Marshal tries to serialize obj parameter
func (j JsonMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("nil object to serialize from")
	}

	return json.Marshal(obj)
}

// Unmarshal tries to deserialize input buffer values into input object
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

// IsInterfaceNil returns true if there is no value under the interface
func (j *JsonMarshalizer) IsInterfaceNil() bool {
	if j == nil {
		return true
	}
	return false
}
