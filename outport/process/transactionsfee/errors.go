package transactionsfee

import "errors"

// ErrNilTransactionFeeCalculator signals that a nil transaction fee calculator has been provided
var ErrNilTransactionFeeCalculator = errors.New("nil transaction fee calculator")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilStorage signals that a nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil Marshaller")
