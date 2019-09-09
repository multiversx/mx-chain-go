package hooks

import "errors"

// ErrNotImplemented signals that a functionality can not be used as it is not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrEmptyCode signals that an account does not contain code
var ErrEmptyCode = errors.New("empty code in provided smart contract holding account")

// ErrAddressLengthNotCorrect signals that an account does not have the correct address
var ErrAddressLengthNotCorrect = errors.New("address length is not correct")

// ErrAddressIsInUnknownShard signals that an addresses computed shard id is unknown
var ErrAddressIsInUnknownShard = errors.New("address is in unknown shard")
