package marshaling

import (
	"encoding/json"
)

var _ Marshalizer = (*jsonMarshalizer)(nil)

type jsonMarshalizer struct {
}

func (marshalizer *jsonMarshalizer) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (marshalizer *jsonMarshalizer) Unmarshal(data interface{}, dataBytes []byte) error {
	return json.Unmarshal(dataBytes, data)
}

func (marshalizer *jsonMarshalizer) IsInterfaceNil() bool {
	return marshalizer == nil
}
