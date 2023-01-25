package poolsCleaner

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const minRoundsToKeepUnprocessedData = int64(1)

// ArgBasePoolsCleaner is the base argument structure used to create pools cleaners
type ArgBasePoolsCleaner struct {
	RoundHandler                   process.RoundHandler
	ShardCoordinator               sharding.Coordinator
	MaxRoundsToKeepUnprocessedData int64
}

type basePoolsCleaner struct {
	roundHandler                   process.RoundHandler
	shardCoordinator               sharding.Coordinator
	maxRoundsToKeepUnprocessedData int64
	cancelFunc                     func()
	isCleaningRoutineRunning       bool
	mut                            sync.Mutex
}

func newBasePoolsCleaner(args ArgBasePoolsCleaner) basePoolsCleaner {
	return basePoolsCleaner{
		roundHandler:                   args.RoundHandler,
		shardCoordinator:               args.ShardCoordinator,
		maxRoundsToKeepUnprocessedData: args.MaxRoundsToKeepUnprocessedData,
	}
}

func checkBaseArgs(args ArgBasePoolsCleaner) error {
	if check.IfNil(args.RoundHandler) {
		return process.ErrNilRoundHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if args.MaxRoundsToKeepUnprocessedData < minRoundsToKeepUnprocessedData {
		return fmt.Errorf("%w for MaxRoundsToKeepUnprocessedData, received %d, min expected %d",
			process.ErrInvalidValue, args.MaxRoundsToKeepUnprocessedData, minRoundsToKeepUnprocessedData)
	}

	return nil
}

// Close will close the endless running go routine
func (base *basePoolsCleaner) Close() error {
	base.mut.Lock()
	defer base.mut.Unlock()

	if base.cancelFunc != nil {
		base.cancelFunc()
		base.isCleaningRoutineRunning = false
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (base *basePoolsCleaner) IsInterfaceNil() bool {
	return base == nil
}
