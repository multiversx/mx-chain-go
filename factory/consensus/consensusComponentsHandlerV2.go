package consensus

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
)

var _ factory.ComponentHandler = (*managedConsensusComponentsV2)(nil)
var _ factory.ConsensusComponentsHolder = (*managedConsensusComponentsV2)(nil)
var _ factory.ConsensusComponentsHandler = (*managedConsensusComponentsV2)(nil)

type managedConsensusComponentsV2 struct {
	*managedConsensusComponents
	consensusComponentsFactoryV2 *consensusComponentsFactoryV2
}

// NewManagedConsensusComponentsV2 creates a managed consensus components handler
func NewManagedConsensusComponentsV2(
	managedConsensusComponents *managedConsensusComponents,
	ccf *consensusComponentsFactoryV2,
) (*managedConsensusComponentsV2, error) {
	if managedConsensusComponents == nil {
		return nil, errors.ErrNilManagedConsensusComponents
	}
	if ccf == nil {
		return nil, errors.ErrNilConsensusComponentsFactory
	}

	return &managedConsensusComponentsV2{
		managedConsensusComponents:   managedConsensusComponents,
		consensusComponentsFactoryV2: ccf,
	}, nil
}

// Create creates the consensus components
func (mcc *managedConsensusComponentsV2) Create() error {
	cc, err := mcc.consensusComponentsFactoryV2.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConsensusComponentsFactoryCreate, err)
	}

	mcc.mutConsensusComponents.Lock()
	mcc.consensusComponents = cc
	mcc.mutConsensusComponents.Unlock()

	return nil
}
