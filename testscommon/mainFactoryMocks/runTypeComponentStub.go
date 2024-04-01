package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-go/consensus"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	processGenesis "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	testFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	BlockChainHookHandlerFactory        hooks.BlockChainHookHandlerCreator
	BlockProcessorFactory               block.BlockProcessorCreator
	BlockTrackerFactory                 track.BlockTrackerCreator
	BootstrapperFromStorageFactory      storageBootstrap.BootstrapperFromStorageCreator
	BootstrapperFactory                 storageBootstrap.BootstrapperCreator
	EpochStartBootstrapperFactory       bootstrap.EpochStartBootstrapperCreator
	ForkDetectorFactory                 sync.ForkDetectorCreator
	HeaderValidatorFactory              block.HeaderValidatorCreator
	RequestHandlerFactory               requestHandlers.RequestHandlerCreator
	ScheduledTxsExecutionFactory        preprocess.ScheduledTxsExecutionCreator
	TransactionCoordinatorFactory       coordinator.TransactionCoordinatorCreator
	ValidatorStatisticsProcessorFactory peer.ValidatorStatisticsProcessorCreator
	AdditionalStorageServiceFactory     process.AdditionalStorageServiceCreator
	SCResultsPreProcessorFactory        preprocess.SmartContractResultPreProcessorCreator
	SCProcessorFactory                  scrCommon.SCProcessorCreator
	ConsensusModelType                  consensus.ConsensusModel
	VmContainerMetaFactory              factoryVm.VmContainerCreator
	VmContainerShardFactory             factoryVm.VmContainerCreator
	AccountCreator                      state.AccountFactory
	OutGoingOperationsPool              sovereignBlock.OutGoingOperationsPool
	DataCodec                           sovereign.DataDecoderHandler
	TopicsChecker                       sovereign.TopicsCheckerHandler
	ShardCoordinatorFactory             sharding.ShardCoordinatorFactory
	GenesisBlockFactory                 processGenesis.GenesisBlockCreatorFactory
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		BlockChainHookHandlerFactory:        &testFactory.BlockChainHookHandlerFactoryMock{},
		BlockProcessorFactory:               &testFactory.BlockProcessorFactoryMock{},
		BlockTrackerFactory:                 &testFactory.BlockTrackerFactoryMock{},
		BootstrapperFromStorageFactory:      &testFactory.BootstrapperFromStorageFactoryMock{},
		BootstrapperFactory:                 &testFactory.BootstrapperFactoryMock{},
		EpochStartBootstrapperFactory:       &testFactory.EpochStartBootstrapperFactoryMock{},
		ForkDetectorFactory:                 &testFactory.ForkDetectorFactoryMock{},
		HeaderValidatorFactory:              &testFactory.HeaderValidatorFactoryMock{},
		RequestHandlerFactory:               &testFactory.RequestHandlerFactoryMock{},
		ScheduledTxsExecutionFactory:        &testFactory.ScheduledTxsExecutionFactoryMock{},
		TransactionCoordinatorFactory:       &testFactory.TransactionCoordinatorFactoryMock{},
		ValidatorStatisticsProcessorFactory: &testFactory.ValidatorStatisticsProcessorFactoryMock{},
		AdditionalStorageServiceFactory:     &testFactory.AdditionalStorageServiceFactoryMock{},
		SCResultsPreProcessorFactory:        &testFactory.SmartContractResultPreProcessorFactoryMock{},
		SCProcessorFactory:                  &testFactory.SCProcessorFactoryMock{},
		ConsensusModelType:                  consensus.ConsensusModelV1,
		VmContainerMetaFactory:              &testFactory.VMContainerMetaFactoryMock{},
		VmContainerShardFactory:             &testFactory.VMContainerShardFactoryMock{},
		AccountCreator:                      &stateMock.AccountsFactoryStub{},
		OutGoingOperationsPool:              &sovereignMocks.OutGoingOperationsPoolMock{},
		DataCodec:                           &sovereignMocks.DataCodecMock{},
		TopicsChecker:                       &sovereignMocks.TopicsCheckerMock{},
		ShardCoordinatorFactory:             sharding.NewMultiShardCoordinatorFactory(),
		GenesisBlockFactory:                 &testFactory.GenesisBlockCreatorFactoryMock{},
	}
}

// Create -
func (r *RunTypeComponentsStub) Create() error {
	return nil
}

// Close -
func (r *RunTypeComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (r *RunTypeComponentsStub) CheckSubcomponents() error {
	return nil
}

// BlockChainHookHandlerCreator -
func (r *RunTypeComponentsStub) String() string {
	return ""
}

// BlockChainHookHandlerCreator -
func (r *RunTypeComponentsStub) BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator {
	return r.BlockChainHookHandlerFactory
}

// BlockProcessorCreator -
func (r *RunTypeComponentsStub) BlockProcessorCreator() block.BlockProcessorCreator {
	return r.BlockProcessorFactory
}

// BlockTrackerCreator -
func (r *RunTypeComponentsStub) BlockTrackerCreator() track.BlockTrackerCreator {
	return r.BlockTrackerFactory
}

// BootstrapperFromStorageCreator -
func (r *RunTypeComponentsStub) BootstrapperFromStorageCreator() storageBootstrap.BootstrapperFromStorageCreator {
	return r.BootstrapperFromStorageFactory
}

// BootstrapperCreator -
func (r *RunTypeComponentsStub) BootstrapperCreator() storageBootstrap.BootstrapperCreator {
	return r.BootstrapperFactory
}

// EpochStartBootstrapperCreator -
func (r *RunTypeComponentsStub) EpochStartBootstrapperCreator() bootstrap.EpochStartBootstrapperCreator {
	return r.EpochStartBootstrapperFactory
}

// ForkDetectorCreator -
func (r *RunTypeComponentsStub) ForkDetectorCreator() sync.ForkDetectorCreator {
	return r.ForkDetectorFactory
}

// HeaderValidatorCreator -
func (r *RunTypeComponentsStub) HeaderValidatorCreator() block.HeaderValidatorCreator {
	return r.HeaderValidatorFactory
}

// RequestHandlerCreator -
func (r *RunTypeComponentsStub) RequestHandlerCreator() requestHandlers.RequestHandlerCreator {
	return r.RequestHandlerFactory
}

// ScheduledTxsExecutionCreator -
func (r *RunTypeComponentsStub) ScheduledTxsExecutionCreator() preprocess.ScheduledTxsExecutionCreator {
	return r.ScheduledTxsExecutionFactory
}

// TransactionCoordinatorCreator -
func (r *RunTypeComponentsStub) TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator {
	return r.TransactionCoordinatorFactory
}

// ValidatorStatisticsProcessorCreator -
func (r *RunTypeComponentsStub) ValidatorStatisticsProcessorCreator() peer.ValidatorStatisticsProcessorCreator {
	return r.ValidatorStatisticsProcessorFactory
}

// AdditionalStorageServiceCreator -
func (r *RunTypeComponentsStub) AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator {
	return r.AdditionalStorageServiceFactory
}

// SCProcessorCreator -
func (r *RunTypeComponentsStub) SCProcessorCreator() scrCommon.SCProcessorCreator {
	return r.SCProcessorFactory
}

// SCResultsPreProcessorCreator -
func (r *RunTypeComponentsStub) SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator {
	return r.SCResultsPreProcessorFactory
}

// ConsensusModel -
func (r *RunTypeComponentsStub) ConsensusModel() consensus.ConsensusModel {
	return r.ConsensusModelType
}

// VmContainerMetaFactoryCreator -
func (r *RunTypeComponentsStub) VmContainerMetaFactoryCreator() factoryVm.VmContainerCreator {
	return r.VmContainerMetaFactory
}

// VmContainerShardFactoryCreator -
func (r *RunTypeComponentsStub) VmContainerShardFactoryCreator() factoryVm.VmContainerCreator {
	return r.VmContainerShardFactory
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
func (r *RunTypeComponentsStub) DataCodecHandler() sovereign.DataDecoderHandler {
	return r.DataCodec
}

// TopicsCheckerHandler -
func (r *RunTypeComponentsStub) TopicsCheckerHandler() sovereign.TopicsCheckerHandler {
	return r.TopicsChecker
}

// ShardCoordinatorCreator -
func (r *RunTypeComponentsStub) ShardCoordinatorCreator() sharding.ShardCoordinatorFactory {
	return r.ShardCoordinatorFactory
}

// ShardCoordinatorCreator -
func (r *RunTypeComponentsStub) GenesisBlockCreator() processGenesis.GenesisBlockCreatorFactory {
	return r.GenesisBlockFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
