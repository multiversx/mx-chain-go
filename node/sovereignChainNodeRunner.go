package node

import (
	"fmt"

	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	consensusComp "github.com/ElrondNetwork/elrond-go/factory/consensus"
	processComp "github.com/ElrondNetwork/elrond-go/factory/processing"
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

	scnr.createManagedConsensusComponentsMethod = scnr.createManagedConsensusComponents
	scnr.createManagedProcessComponentsMethod = scnr.createManagedProcessComponents

	return scnr, nil
}

func (scnr *sovereignChainNodeRunner) createManagedConsensusComponents(consensusArgs consensusComp.ConsensusComponentsFactoryArgs) (mainFactory.ConsensusComponentsHandler, error) {
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

func (scnr *sovereignChainNodeRunner) createManagedProcessComponents(processArgs processComp.ProcessComponentsFactoryArgs) (mainFactory.ProcessComponentsHandler, error) {
	processComponentsFactory, err := processComp.NewProcessComponentsFactory(processArgs)
	if err != nil {
		return nil, fmt.Errorf("NewProcessComponentsFactory failed: %w", err)
	}

	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		return nil, err
	}

	sovereignChainProcessComponentsFactory, err := processComp.NewSovereignChainProcessComponentsFactory(processComponentsFactory)
	if err != nil {
		return nil, fmt.Errorf("NewSovereignChainProcessComponentsFactory failed: %w", err)
	}

	sovereignChainManagedProcessComponents, err := processComp.NewSovereignChainManagedProcessComponents(managedProcessComponents, sovereignChainProcessComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = sovereignChainManagedProcessComponents.Create()
	if err != nil {
		return nil, err
	}

	return sovereignChainManagedProcessComponents, nil
}
