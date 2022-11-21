package errors

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/common"
)

// IsGetNodeFromDBError returns true if the provided error is of type getNodeFromDB
func IsGetNodeFromDBError(err error) bool {
	if err == nil {
		return false
	}

	if IsClosingError(err) {
		return false
	}

	if strings.Contains(err.Error(), common.GetNodeFromDBErrorString) {
		return true
	}

	return false
}

// GetNodeFromDBErr defines a custom error for trie get node
type GetNodeFromDBErr struct {
	key []byte
}

// NewGetNodeFromDBErr will create a new instance of GetNodeFromDBErr
func NewGetNodeFromDBErr(key []byte) *GetNodeFromDBErr {
	return &GetNodeFromDBErr{key: key}
}

// Error returns the error as string
func (e *GetNodeFromDBErr) Error() string {
	return fmt.Sprintf(
		"%s for key %v",
		common.GetNodeFromDBErrorString,
		hex.EncodeToString(e.key),
	)
}

// GetKey will return the key that generated the error
func (e *GetNodeFromDBErr) GetKey() []byte {
	return e.key
}
