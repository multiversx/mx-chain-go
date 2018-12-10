package marshal

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type JsonMarshalizer struct {
}

func (j JsonMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("NIL object to serilize from!")
	}

	return json.Marshal(obj)
}

func (j JsonMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
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
