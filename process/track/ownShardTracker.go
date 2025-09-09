package track

import (
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

const minNonceDifference = 5

type ownShardTracker struct {
	enableEpochsHandler common.EnableEpochsHandler
	// maximum difference between the nonce of the last notarized execution result and the current header nonce
	maxNonceDifference uint64
	ownShardStuck      atomic.Bool
}

// NewOwnShardTracker creates a new instance of ownShardTracker
func NewOwnShardTracker(enableEpochsHandler common.EnableEpochsHandler, maxNonceDifference uint64) (*ownShardTracker, error) {
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if maxNonceDifference < minNonceDifference {
		return nil, process.ErrInvalidMaxNonceDifference
	}

	return &ownShardTracker{
		enableEpochsHandler: enableEpochsHandler,
		maxNonceDifference:  maxNonceDifference,
		ownShardStuck:       atomic.Bool{},
	}, nil
}

// ComputeOwnShardStuck computes if the own shard is stuck based on the last execution results info and the current nonce.
func (ost *ownShardTracker) ComputeOwnShardStuck(lastExecutionResultsInfo data.BaseExecutionResultHandler, currentNonce uint64) {
	if !ost.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return
	}

	lastNotarizedNonce := lastExecutionResultsInfo.GetHeaderNonce()
	if currentNonce > lastNotarizedNonce && currentNonce-lastNotarizedNonce > ost.maxNonceDifference {
		ost.ownShardStuck.Store(true)
		return
	}
	ost.ownShardStuck.Store(false)
}

// IsOwnShardStuck returns true if the own shard is stuck, false otherwise
func (ost *ownShardTracker) IsOwnShardStuck() bool {
	return ost.ownShardStuck.Load()
}
