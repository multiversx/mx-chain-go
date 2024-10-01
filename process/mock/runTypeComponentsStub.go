package mock

import (
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	storageRequestFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	syncerFactory "github.com/multiversx/mx-chain-go/state/syncer/factory"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AdditionalStorageServiceFactory             process.AdditionalStorageServiceCreator
	ShardCoordinatorFactory                     sharding.ShardCoordinatorFactory
	NodesCoordinatorWithRaterFactory            nodesCoordinator.NodesCoordinatorWithRaterFactory
	RequestHandlerFactory                       requestHandlers.RequestHandlerCreator
	AccountCreator                              state.AccountFactory
	OutGoingOperationsPool                      sovereignBlock.OutGoingOperationsPool
	DataCodec                                   sovereign.DataCodecHandler
	TopicsChecker                               sovereign.TopicsCheckerHandler
	RequestersContainerFactoryCreatorField      requesterscontainer.RequesterContainerFactoryCreator
	ValidatorAccountsSyncerFactoryHandlerField  syncerFactory.ValidatorAccountsSyncerFactoryHandler
	ShardRequestersContainerCreatorHandlerField storageRequestFactory.ShardRequestersContainerCreatorHandler
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AdditionalStorageServiceFactory:             &testscommon.AdditionalStorageServiceFactoryMock{},
		ShardCoordinatorFactory:                     sharding.NewMultiShardCoordinatorFactory(),
		NodesCoordinatorWithRaterFactory:            nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		RequestHandlerFactory:                       requestHandlers.NewResolverRequestHandlerFactory(),
		AccountCreator:                              &stateMock.AccountsFactoryStub{},
		OutGoingOperationsPool:                      &sovereignMocks.OutGoingOperationsPoolMock{},
		DataCodec:                                   &sovereignMocks.DataCodecMock{},
		TopicsChecker:                               &sovereignMocks.TopicsCheckerMock{},
		RequestersContainerFactoryCreatorField:      requesterscontainer.NewShardRequestersContainerFactoryCreator(),
		ValidatorAccountsSyncerFactoryHandlerField:  syncerFactory.NewValidatorAccountsSyncerFactory(),
		ShardRequestersContainerCreatorHandlerField: storageRequestFactory.NewShardRequestersContainerCreator(),
	}
}

// NewSovereignRunTypeComponentsStub -
func NewSovereignRunTypeComponentsStub() *RunTypeComponentsStub {
	rt := NewRunTypeComponentsStub()
	requestHandlerFactory, _ := requestHandlers.NewSovereignResolverRequestHandlerFactory(rt.RequestHandlerFactory)

	return &RunTypeComponentsStub{
		AdditionalStorageServiceFactory:             factory.NewSovereignAdditionalStorageServiceFactory(),
		ShardCoordinatorFactory:                     sharding.NewSovereignShardCoordinatorFactory(),
		NodesCoordinatorWithRaterFactory:            &testscommon.NodesCoordinatorFactoryMock{},
		RequestHandlerFactory:                       requestHandlerFactory,
		AccountCreator:                              &stateMock.AccountsFactoryStub{},
		OutGoingOperationsPool:                      &sovereignMocks.OutGoingOperationsPoolMock{},
		DataCodec:                                   &sovereignMocks.DataCodecMock{},
		TopicsChecker:                               &sovereignMocks.TopicsCheckerMock{},
		RequestersContainerFactoryCreatorField:      requesterscontainer.NewSovereignShardRequestersContainerFactoryCreator(),
		ValidatorAccountsSyncerFactoryHandlerField:  syncerFactory.NewSovereignValidatorAccountsSyncerFactory(),
		ShardRequestersContainerCreatorHandlerField: storageRequestFactory.NewSovereignShardRequestersContainerCreator(),
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

// RequestersContainerFactoryCreator -
func (r *RunTypeComponentsStub) RequestersContainerFactoryCreator() requesterscontainer.RequesterContainerFactoryCreator {
	return r.RequestersContainerFactoryCreatorField
}

// ValidatorAccountsSyncerFactoryHandler -
func (r *RunTypeComponentsStub) ValidatorAccountsSyncerFactoryHandler() syncerFactory.ValidatorAccountsSyncerFactoryHandler {
	return r.ValidatorAccountsSyncerFactoryHandlerField
}

// ShardRequestersContainerCreatorHandler -
func (r *RunTypeComponentsStub) ShardRequestersContainerCreatorHandler() storageRequestFactory.ShardRequestersContainerCreatorHandler {
	return r.ShardRequestersContainerCreatorHandlerField
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
