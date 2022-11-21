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
	getErr error
	key    []byte
}

// NewGetNodeFromDBErr will create a new instance of GetNodeFromDBErr
func NewGetNodeFromDBErr(key []byte, err error) *GetNodeFromDBErr {
	return &GetNodeFromDBErr{getErr: err, key: key}
}

// Error returns the error as string
func (e *GetNodeFromDBErr) Error() string {
	return fmt.Sprintf(
		"%s: %s for key %v",
		common.GetNodeFromDBErrorString,
		e.getErr.Error(),
		hex.EncodeToString(e.key),
	)
}

// GetKey will return the key that generated the error
func (e *GetNodeFromDBErr) GetKey() []byte {
	return e.key
}
