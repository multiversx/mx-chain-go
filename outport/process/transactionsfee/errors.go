package transactionsfee

import "errors"

// ErrNilTransactionFeeCalculator signals that a nil transaction fee calculator has been provided
var ErrNilTransactionFeeCalculator = errors.New("nil transaction fee calculator")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilStorage signals that a nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")
