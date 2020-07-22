package smartContract

import (
	"errors"
)

// ErrUpgradeNotAllowed signals that upgrade is not allowed
var ErrUpgradeNotAllowed = errors.New("upgrade not allowed")
