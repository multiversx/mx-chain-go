package metachain

import "errors"

var errNilValidatorsInfoMap = errors.New("received nil shard validators info map")

var errCannotComputeDenominator = errors.New("cannot compute denominator value")
