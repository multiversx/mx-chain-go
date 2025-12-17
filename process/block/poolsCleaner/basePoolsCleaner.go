package poolsCleaner

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgBasePoolsCleaner is the base argument structure used to create pools cleaners
type ArgBasePoolsCleaner struct {
	RoundHandler          process.RoundHandler
	ShardCoordinator      sharding.Coordinator
	ProcessConfigsHandler common.ProcessConfigsHandler
}

type basePoolsCleaner struct {
	roundHandler             process.RoundHandler
	shardCoordinator         sharding.Coordinator
	processConfigsHandler    common.ProcessConfigsHandler
	cancelFunc               func()
	isCleaningRoutineRunning bool
	mut                      sync.Mutex
}

func newBasePoolsCleaner(args ArgBasePoolsCleaner) basePoolsCleaner {
	return basePoolsCleaner{
		roundHandler:          args.RoundHandler,
		shardCoordinator:      args.ShardCoordinator,
		processConfigsHandler: args.ProcessConfigsHandler,
	}
}

func checkBaseArgs(args ArgBasePoolsCleaner) error {
	if check.IfNil(args.RoundHandler) {
		return process.ErrNilRoundHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ProcessConfigsHandler) {
		return process.ErrNilProcessConfigsHandler
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
