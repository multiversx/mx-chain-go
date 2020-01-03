package headerCheck

import "errors"

// ErrNotEnoughSignatures signals that a block is not signed by at least the minimum number of validators from
// the consensus group
var ErrNotEnoughSignatures = errors.New("not enough signatures in block")
