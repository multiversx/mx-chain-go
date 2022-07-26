package transactionsfee

import "errors"

// errNilTransactionFeeCalculator signals that a nil transaction fee calculator has been provided
var errNilTransactionFeeCalculator = errors.New("nil transaction fee calculator")

// errNilShardCoordinator signals that a nil shard coordinator has been provided
var errNilShardCoordinator = errors.New("nil shard coordinator")

// errNilStorage signals that a nil storage has been provided
var errNilStorage = errors.New("nil storage")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var errNilMarshalizer = errors.New("nil Marshalizer")
