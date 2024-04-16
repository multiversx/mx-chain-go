package mock

import (
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AdditionalStorageServiceFactory  process.AdditionalStorageServiceCreator
	ShardCoordinatorFactory          sharding.ShardCoordinatorFactory
	NodesCoordinatorWithRaterFactory nodesCoordinator.NodesCoordinatorWithRaterFactory
	AccountCreator                   state.AccountFactory
	OutGoingOperationsPool           sovereignBlock.OutGoingOperationsPool
	DataCodec                        sovereign.DataCodecHandler
	TopicsChecker                    sovereign.TopicsCheckerHandler
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AdditionalStorageServiceFactory:  &testscommon.AdditionalStorageServiceFactoryMock{},
		ShardCoordinatorFactory:          sharding.NewMultiShardCoordinatorFactory(),
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		AccountCreator:                   &stateMock.AccountsFactoryStub{},
		OutGoingOperationsPool:           &sovereignMocks.OutGoingOperationsPoolMock{},
		DataCodec:                        &sovereignMocks.DataCodecMock{},
		TopicsChecker:                    &sovereignMocks.TopicsCheckerMock{},
	}
}

// AdditionalStorageServiceCreator -
func (r *RunTypeComponentsStub) AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator {
	return r.AdditionalStorageServiceFactory
}

// ShardCoordinatorCreator -
func (r *RunTypeComponentsStub) ShardCoordinatorCreator() sharding.ShardCoordinatorFactory {
	return r.ShardCoordinatorFactory
}

// NodesCoordinatorWithRaterCreator -
func (r *RunTypeComponentsStub) NodesCoordinatorWithRaterCreator() nodesCoordinator.NodesCoordinatorWithRaterFactory {
	return r.NodesCoordinatorWithRaterFactory
}

// AccountsCreator -
func (r *RunTypeComponentsStub) AccountsCreator() state.AccountFactory {
	return r.AccountCreator
}

// OutGoingOperationsPoolHandler -
func (r *RunTypeComponentsStub) OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool {
	return r.OutGoingOperationsPool
}

// DataCodecHandler -
func (r *RunTypeComponentsStub) DataCodecHandler() sovereign.DataCodecHandler {
	return r.DataCodec
}

// TopicsCheckerHandler -
func (r *RunTypeComponentsStub) TopicsCheckerHandler() sovereign.TopicsCheckerHandler {
	return r.TopicsChecker
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
