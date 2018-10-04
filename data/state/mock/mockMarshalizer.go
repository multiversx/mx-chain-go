package mock

import (
	"encoding/json"
	"errors"
)

type MockMarshalizer struct {
}

func (*MockMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("NIL object to serilize from!")
	}

	return json.Marshal(obj)
}

func (*MockMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
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

func (*MockMarshalizer) Version() string {
	return "JSON/v.0.0.0.1"
}
