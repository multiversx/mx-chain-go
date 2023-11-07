package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
)

// ScheduledTxsExecutionFactoryArgs holds all dependencies required by the process data factory to create components
type ScheduledTxsExecutionFactoryArgs struct {
	TxProcessor             process.TransactionProcessor
	TxCoordinator           process.TransactionCoordinator
	Storer                  storage.Storer
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	ShardCoordinator        sharding.Coordinator
	TxExecutionOrderHandler common.TxExecutionOrderHandler
}

type shardScheduledTxsExecutionFactory struct {
}

// NewShardScheduledTxsExecutionFactory creates a new shard scheduled txs execution factory
func NewShardScheduledTxsExecutionFactory() (*shardScheduledTxsExecutionFactory, error) {
	return &shardScheduledTxsExecutionFactory{}, nil
}

// CreateScheduledTxsExecutionHandler creates a new scheduled txs execution handler for shard chain
func (stxef *shardScheduledTxsExecutionFactory) CreateScheduledTxsExecutionHandler(args ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	return NewScheduledTxsExecution(
		args.TxProcessor,
		args.TxCoordinator,
		args.Storer,
		args.Marshalizer,
		args.Hasher,
		args.ShardCoordinator,
		args.TxExecutionOrderHandler,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stxef *shardScheduledTxsExecutionFactory) IsInterfaceNil() bool {
	return stxef == nil
}
