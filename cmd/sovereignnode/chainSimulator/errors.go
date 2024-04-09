package chainSimulator

import "errors"

var (
	errNilChainSimulator     = errors.New("nil chain Simulator")
	errNilMetachainNode      = errors.New("nil metachain node")
	errShardSetupError       = errors.New("shard setup error")
	errEmptySliceOfTxs       = errors.New("empty slice of transactions to send")
	errNilTransaction        = errors.New("nil transaction")
	errInvalidMaxNumOfBlocks = errors.New("invalid max number of blocks to generate")
)
