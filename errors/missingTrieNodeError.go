package errors

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
)

// GetNodeFromDBErrWithKey defines a custom error for trie get node
type GetNodeFromDBErrWithKey struct {
	getErr       error
	key          []byte
	dbIdentifier string
}

// NewGetNodeFromDBErrWithKey will create a new instance of GetNodeFromDBErrWithKey
func NewGetNodeFromDBErrWithKey(key []byte, err error, id string) *GetNodeFromDBErrWithKey {
	return &GetNodeFromDBErrWithKey{
		getErr:       err,
		key:          key,
		dbIdentifier: id,
	}
}

// Error returns the error as string
func (e *GetNodeFromDBErrWithKey) Error() string {
	return fmt.Sprintf(
		"%s: %s for key %v",
		core.GetNodeFromDBErrorString,
		e.getErr.Error(),
		hex.EncodeToString(e.key),
	)
}

// GetKey will return the key that generated the error
func (e *GetNodeFromDBErrWithKey) GetKey() []byte {
	return e.key
}

// GetIdentifier will return the db identifier corresponding to the db
func (e *GetNodeFromDBErrWithKey) GetIdentifier() string {
	return e.dbIdentifier
}
