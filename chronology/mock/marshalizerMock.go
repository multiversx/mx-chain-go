package mock

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type MarshalizerMock struct {
}

func (m *MarshalizerMock) Marshal(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("NIL object to serilize from!")
	}

	return json.Marshal(obj)
}

func (m *MarshalizerMock) Unmarshal(obj interface{}, buff []byte) error {
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
