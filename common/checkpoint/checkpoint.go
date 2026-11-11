package checkpoint

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrInvalidCheckpoint signals that the provided checkpoint can not be decoded
var ErrInvalidCheckpoint = errors.New("invalid checkpoint")

// Checkpoint holds what is needed to resume a paginated trie listing: the block the listing was started at
// and the trie iterator state to continue from. It is returned as a JSON object and sent back either as the
// same object in a POST body, or as its base64url-encoded JSON in a GET query parameter
type Checkpoint struct {
	BlockNonce    uint64 `json:"blockNonce"`
	BlockHash     string `json:"blockHash"`
	BlockRootHash string `json:"blockRootHash"`
	// Epoch selects the epoch storage when the listing is resumed by root hash only (no block hash to derive it from)
	Epoch         *uint32  `json:"epoch,omitempty"`
	IteratorState [][]byte `json:"iteratorState"`
}

// Encode returns the checkpoint as base64url-encoded JSON, for use in a query parameter
func Encode(cp *Checkpoint) (string, error) {
	buff, err := json.Marshal(cp)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buff), nil
}

// Decode parses a string produced by Encode
func Decode(encoded string) (*Checkpoint, error) {
	buff, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCheckpoint, err.Error())
	}

	cp := &Checkpoint{}
	err = json.Unmarshal(buff, cp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCheckpoint, err.Error())
	}

	return cp, nil
}
