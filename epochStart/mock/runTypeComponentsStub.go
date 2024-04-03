package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AdditionalStorageServiceFactory  process.AdditionalStorageServiceCreator
	ShardCoordinatorFactory          sharding.ShardCoordinatorFactory
	NodesCoordinatorWithRaterFactory nodesCoordinator.NodesCoordinatorWithRaterFactory
	RequestHandlerFactory            requestHandlers.RequestHandlerCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AdditionalStorageServiceFactory:  &testscommon.AdditionalStorageServiceFactoryMock{},
		ShardCoordinatorFactory:          sharding.NewMultiShardCoordinatorFactory(),
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		RequestHandlerFactory:            requestHandlers.NewResolverRequestHandlerFactory(),
	}
}

// NewSovereignRunTypeComponentsStub -
func NewSovereignRunTypeComponentsStub() *RunTypeComponentsStub {
	runTypeComponents := NewRunTypeComponentsStub()

	additionalStorageServiceCreator, _ := storageFactory.NewSovereignAdditionalStorageServiceFactory()
	requestHandlerFactory, _ := requestHandlers.NewSovereignResolverRequestHandlerFactory(runTypeComponents.RequestHandlerFactory)

	return &RunTypeComponentsStub{
		AdditionalStorageServiceFactory:  additionalStorageServiceCreator,
		ShardCoordinatorFactory:          sharding.NewSovereignShardCoordinatorFactory(),
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewSovereignIndexHashedNodesCoordinatorWithRaterFactory(),
		RequestHandlerFactory:            requestHandlerFactory,
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

// RequestHandlerCreator -
func (r *RunTypeComponentsStub) RequestHandlerCreator() requestHandlers.RequestHandlerCreator {
	return r.RequestHandlerFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
