package metachain

import "errors"

var errNilValidatorsInfoMap = errors.New("received nil shard validators info map")

var errCannotComputeDenominator = errors.New("cannot compute denominator value")

var errNilAuctionListDisplayHandler = errors.New("nil auction list display handler provided")

var errNilTableDisplayHandler = errors.New("nil table display handler provided")

var errHashMismatch = errors.New("hash mismatch")

var errMethodNotSupported = errors.New("method not supported")
