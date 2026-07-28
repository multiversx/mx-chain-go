package errors

import "errors"

// ErrEmptySliceOfTxs signals that an empty slice of transactions has been provided
var ErrEmptySliceOfTxs = errors.New("empty slice of transactions to send")

// ErrNilTransaction signals that a nil transaction has been provided
var ErrNilTransaction = errors.New("nil transaction")

// ErrInvalidMaxNumOfBlocks signals that an invalid max numerof blocks has been provided
var ErrInvalidMaxNumOfBlocks = errors.New("invalid max number of blocks to generate")

// ErrInvalidConsensusMode signals that an unsupported chain-simulator consensus mode was provided.
var ErrInvalidConsensusMode = errors.New("invalid chain simulator consensus mode")
