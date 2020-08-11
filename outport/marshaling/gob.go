package marshaling

import (
	"bytes"
	"encoding/gob"
)

var _ Marshalizer = (*gobMarshalizer)(nil)

type gobMarshalizer struct {
}

func (marshalizer *gobMarshalizer) Marshal(data interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (marshalizer *gobMarshalizer) Unmarshal(data interface{}, dataBytes []byte) error {
	buffer := bytes.NewBuffer(dataBytes)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (marshalizer *gobMarshalizer) IsInterfaceNil() bool {
	return marshalizer == nil
}
