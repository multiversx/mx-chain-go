package node

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	consensusComp "github.com/ElrondNetwork/elrond-go/factory/consensus"
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
	coreComponents mainFactory.CoreComponentsHolder,
	networkComponents mainFactory.NetworkComponentsHolder,
	cryptoComponents mainFactory.CryptoComponentsHolder,
	dataComponents mainFactory.DataComponentsHolder,
	stateComponents mainFactory.StateComponentsHolder,
	statusComponents mainFactory.StatusComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
) (mainFactory.ConsensusComponentsHandler, error) {
	scheduledProcessorArgs := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}

	scheduledProcessor, err := spos.NewScheduledProcessorWrapper(scheduledProcessorArgs)
	if err != nil {
		return nil, err
	}

	consensusArgs := consensusComp.ConsensusComponentsFactoryArgs{
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

	consensusFactory, err := consensusComp.NewConsensusComponentsFactory(consensusArgs)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
	}

	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
	if err != nil {
		return nil, err
	}

	consensusFactoryV2, err := consensusComp.NewConsensusComponentsFactoryV2(consensusFactory)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusComponentsFactoryV2 failed: %w", err)
	}

	managedConsensusComponentsV2, err := consensusComp.NewManagedConsensusComponentsV2(managedConsensusComponents, consensusFactoryV2)
	if err != nil {
		return nil, err
	}

	err = managedConsensusComponentsV2.Create()
	if err != nil {
		return nil, err
	}

	return managedConsensusComponentsV2, nil
}
