package marshal

import (
	"bytes"
	"encoding/json"
)

// TxJsonMarshalizer represents a custom marshalizer for transactions which does not escape HTML characters in strings
type TxJsonMarshalizer struct {
}

// Marshal tries to serialize the obj parameter
func (t *TxJsonMarshalizer) Marshal(obj interface{}) ([]byte, error) {
	bytesBuffer := new(bytes.Buffer)
	jsonEncoder := json.NewEncoder(bytesBuffer)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(obj)
	if err != nil {
		return nil, err
	}

	encodedResult := bytesBuffer.Bytes()
	return trimLineFeed(encodedResult), nil
}

func trimLineFeed(bytes []byte) []byte {
	// this should be replaced, but for some reason, bytes.TrimRight(b, "\r") does not work
	lastByte := bytes[len(bytes)-1:]
	if lastByte[0] == byte(10) { // hardcoded for now
		return bytes[:len(bytes)-1]
	}

	return bytes
}

// Unmarshal tries to deserialize input buffer values into input object
func (t *TxJsonMarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	return json.Unmarshal(buff, obj)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *TxJsonMarshalizer) IsInterfaceNil() bool {
	return t == nil
}
