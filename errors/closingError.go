package errors

import (
	"strings"

	"github.com/ElrondNetwork/elrond-go-storage/common"
)

// IsClosingError returns true if the provided error is used whenever the node is in the closing process
func IsClosingError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), common.ErrDBIsClosed.Error()) ||
		strings.Contains(err.Error(), ErrContextClosing.Error())
}
