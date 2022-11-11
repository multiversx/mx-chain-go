package node

import (
	"fmt"

	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
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

	scnr.createManagedProcessComponentsMethod = scnr.createManagedProcessComponents

	return scnr, nil
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
