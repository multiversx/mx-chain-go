package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func encode(obj objectsMap) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(obj); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding : %w", err)
	}

	return buff, nil
}

func getDocumentsByIDsQuery(hashes []string) objectsMap {
	interfaceSlice := make([]interface{}, len(hashes))
	for idx := range hashes {
		interfaceSlice[idx] = objectsMap{
			"_id":     hashes[idx],
			"_source": false,
		}
	}

	return objectsMap{
		"docs": interfaceSlice,
	}
}

func prepareHashesForBulkRemove(hashes []string) objectsMap {
	return objectsMap{
		"query": objectsMap{
			"ids": objectsMap{
				"type":   "_doc",
				"values": hashes,
			},
		},
	}
}
