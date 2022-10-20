package processing

import (
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
)

var _ factory.ComponentHandler = (*sovereignChainManagedProcessComponents)(nil)
var _ factory.ProcessComponentsHolder = (*sovereignChainManagedProcessComponents)(nil)
var _ factory.ProcessComponentsHandler = (*sovereignChainManagedProcessComponents)(nil)

type sovereignChainManagedProcessComponents struct {
	*managedProcessComponents
	sovereignChainProcessComponentsFactory *sovereignChainProcessComponentsFactory
}

// NewSovereignChainManagedProcessComponents returns a news instance of sovereignChainManagedProcessComponents
func NewSovereignChainManagedProcessComponents(
	managedProcessComponents *managedProcessComponents,
	scpcf *sovereignChainProcessComponentsFactory,
) (*sovereignChainManagedProcessComponents, error) {
	if managedProcessComponents == nil {
		return nil, errors.ErrNilManagedProcessComponents
	}
	if scpcf == nil {
		return nil, errors.ErrNilProcessComponentsFactory
	}

	return &sovereignChainManagedProcessComponents{
		managedProcessComponents:               managedProcessComponents,
		sovereignChainProcessComponentsFactory: scpcf,
	}, nil
}

// Create will create the managed components
func (scmpc *sovereignChainManagedProcessComponents) Create() error {
	pc, err := scmpc.sovereignChainProcessComponentsFactory.Create()
	if err != nil {
		return err
	}

	scmpc.mutProcessComponents.Lock()
	scmpc.processComponents = pc
	scmpc.mutProcessComponents.Unlock()

	return nil
}
