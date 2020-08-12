package marshaling

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// SerializableMapStringBytes is a serializable map
type SerializableMapStringBytes struct {
	Keys   [][]byte
	Values [][]byte
}

// NewSerializableMapStringBytes creates a serializable map
func NewSerializableMapStringBytes(input map[string][]byte) *SerializableMapStringBytes {
	s := &SerializableMapStringBytes{
		Keys:   make([][]byte, 0, len(input)),
		Values: make([][]byte, 0, len(input)),
	}

	for key, value := range input {
		s.Keys = append(s.Keys, []byte(key))
		s.Values = append(s.Values, value)
	}

	return s
}

// SerializableMapStringTransactionHandler is a serializable map
type SerializableMapStringTransactionHandler struct {
	Keys   [][]byte
	Values []data.TransactionHandler
}

// NewSerializableMapStringTransactionHandler creates a serializable map
func NewSerializableMapStringTransactionHandler(input map[string]data.TransactionHandler) *SerializableMapStringTransactionHandler {
	s := &SerializableMapStringTransactionHandler{
		Keys:   make([][]byte, 0, len(input)),
		Values: make([]data.TransactionHandler, 0, len(input)),
	}

	for key, value := range input {
		s.Keys = append(s.Keys, []byte(key))
		s.Values = append(s.Values, value)
	}

	return s
}
