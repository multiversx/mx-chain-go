package indexer

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type object = map[string]interface{}

func encode(obj object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(obj); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding : %w", err)
	}

	return buff, nil
}

func getDocumentsByIDsQuery(hashes []string) object {
	interfaceSlice := make([]interface{}, len(hashes))
	for idx := range hashes {
		interfaceSlice[idx] = object{
			"_id":     hashes[idx],
			"_source": false,
		}
	}

	return object{
		"docs": interfaceSlice,
	}
}
