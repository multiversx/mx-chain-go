package chainSimulator

import "errors"

var errNilChainSimulator = errors.New("nil chain simulator")
var errNilMetachainNode = errors.New("nil metachain node")
var errShardSetupError = errors.New("shard setup error")

// ErrEmptySliceOfTxs signals that an empty slice of transactions has been provided
var ErrEmptySliceOfTxs = errors.New("empty slice of transactions to send")

// ErrNilTransaction signals that a nil transaction has been provided
var ErrNilTransaction = errors.New("nil transaction")

// ErrInvalidMaxNumOfBlocks signals that an invalid max numerof blocks has been provided
var ErrInvalidMaxNumOfBlocks = errors.New("invalid max number of blocks to generate")
