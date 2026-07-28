package chainSimulator

import "errors"

var (
	errNilChainSimulator = errors.New("nil chain simulator")
	errNilMetachainNode  = errors.New("nil metachain node")
	errShardSetupError   = errors.New("shard setup error")

	errConsensusNodeNotDriveable = errors.New("node does not support manual consensus drive")
)
