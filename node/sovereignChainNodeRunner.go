package node

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/consensus"
)

// sovereignChainNodeRunner holds the sovereign chain node runner configuration and controls running of a node
type sovereignChainNodeRunner struct {
	*nodeRunner
}

// NewSovereignChainNodeRunner creates a sovereignChainNodeRunner instance
func NewSovereignChainNodeRunner(nodeRunner *nodeRunner) (*sovereignChainNodeRunner, error) {
	if nodeRunner == nil {
		return nil, fmt.Errorf("nil node runner")
	}

	scnr := &sovereignChainNodeRunner{
		nodeRunner,
	}

	scnr.CreateManagedConsensusComponentsMethod = scnr.CreateManagedConsensusComponents

	return scnr, nil
}

// CreateManagedConsensusComponents is the managed consensus components factory
func (scnr *sovereignChainNodeRunner) CreateManagedConsensusComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	statusComponents factory.StatusComponentsHolder,
	processComponents factory.ProcessComponentsHolder,
) (factory.ConsensusComponentsHandler, error) {
	scheduledProcessorArgs := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}

	scheduledProcessor, err := spos.NewScheduledProcessorWrapper(scheduledProcessorArgs)
	if err != nil {
		return nil, err
	}

	consensusArgs := consensus.ConsensusComponentsFactoryArgs{
		Config:                *scnr.configs.GeneralConfig,
		BootstrapRoundIndex:   scnr.configs.FlagsConfig.BootstrapRoundIndex,
		CoreComponents:        coreComponents,
		NetworkComponents:     networkComponents,
		CryptoComponents:      cryptoComponents,
		DataComponents:        dataComponents,
		ProcessComponents:     processComponents,
		StateComponents:       stateComponents,
		StatusComponents:      statusComponents,
		ScheduledProcessor:    scheduledProcessor,
		IsInImportMode:        scnr.configs.ImportDbConfig.IsImportDBMode,
		ShouldDisableWatchdog: scnr.configs.FlagsConfig.DisableConsensusWatchdog,
	}

	consensusFactory, err := consensus.NewConsensusComponentsFactory(consensusArgs)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
	}

	managedConsensusComponents, err := consensus.NewManagedConsensusComponents(consensusFactory)
	if err != nil {
		return nil, err
	}

	consensusFactoryV2, err := consensus.NewConsensusComponentsFactoryV2(consensusFactory)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactoryV2 failed: %w", err)
	}

	managedConsensusComponentsV2, err := consensus.NewManagedConsensusComponentsV2(managedConsensusComponents, consensusFactoryV2)
	if err != nil {
		return nil, err
	}

	err = managedConsensusComponentsV2.Create()
	if err != nil {
		return nil, err
	}

	return managedConsensusComponentsV2, nil
}
