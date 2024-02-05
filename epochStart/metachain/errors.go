package metachain

import "errors"

var errNilValidatorsInfoMap = errors.New("received nil shard validators info map")

var errCannotComputeDenominator = errors.New("cannot compute denominator value")

var errNilAuctionListDisplayHandler = errors.New("nil auction list display handler provided")
