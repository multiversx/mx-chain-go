package transactionsfee

import "errors"

// ErrNilTransactionFeeCalculator signals that a nil transaction fee calculator has been provided
var ErrNilTransactionFeeCalculator = errors.New("nil transaction fee calculator")
