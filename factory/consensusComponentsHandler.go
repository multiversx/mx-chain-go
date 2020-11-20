package factory

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var _ ComponentHandler = (*managedConsensusComponents)(nil)
var _ ConsensusComponentsHolder = (*managedConsensusComponents)(nil)
var _ ConsensusComponentsHandler = (*managedConsensusComponents)(nil)

type managedConsensusComponents struct {
	*consensusComponents
	consensusComponentsFactory *consensusComponentsFactory
	mutConsensusComponents     sync.RWMutex
}

// NewManagedConsensusComponents creates a managed consensus components handler
func NewManagedConsensusComponents(ccf *consensusComponentsFactory) (*managedConsensusComponents, error) {
	if ccf == nil {
		return nil, errors.ErrNilConsensusComponentsFactory
	}

	return &managedConsensusComponents{
		consensusComponents:        nil,
		consensusComponentsFactory: ccf,
	}, nil
}

// Create creates the consensus components
func (mcc *managedConsensusComponents) Create() error {
	cc, err := mcc.consensusComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConsensusComponentsFactoryCreate, err)
	}

	mcc.mutConsensusComponents.Lock()
	mcc.consensusComponents = cc
	mcc.mutConsensusComponents.Unlock()

	return nil
}

// Close closes all the consensus components
func (mcc *managedConsensusComponents) Close() error {
	mcc.mutConsensusComponents.Lock()
	defer mcc.mutConsensusComponents.Unlock()

	if mcc.consensusComponents != nil {
		err := mcc.consensusComponents.Close()
		if err != nil {
			return err
		}
		mcc.consensusComponents = nil
	}

	return nil
}

// Chronology returns the chronology handler
func (mcc *managedConsensusComponents) Chronology() consensus.ChronologyHandler {
	mcc.mutConsensusComponents.RLock()
	defer mcc.mutConsensusComponents.RUnlock()

	if mcc.consensusComponents == nil {
		return nil
	}

	return mcc.consensusComponents.chronology
}

// ConsensusWorker returns the consensus worker
func (mcc *managedConsensusComponents) ConsensusWorker() ConsensusWorker {
	mcc.mutConsensusComponents.RLock()
	defer mcc.mutConsensusComponents.RUnlock()

	if mcc.consensusComponents == nil {
		return nil
	}

	return mcc.consensusComponents.worker
}

// BroadcastMessenger returns the consensus broadcast messenger
func (mcc *managedConsensusComponents) BroadcastMessenger() consensus.BroadcastMessenger {
	mcc.mutConsensusComponents.RLock()
	defer mcc.mutConsensusComponents.RUnlock()

	if mcc.consensusComponents == nil {
		return nil
	}

	return mcc.consensusComponents.broadcastMessenger
}

// ConsensusGroupSize returns the consensus group size
func (mcc *managedConsensusComponents) ConsensusGroupSize() (int, error) {
	mcc.mutConsensusComponents.RLock()
	defer mcc.mutConsensusComponents.RUnlock()

	if mcc.consensusComponents == nil {
		return 0, errors.ErrNilConsensusComponentsHolder
	}

	return mcc.consensusComponents.consensusGroupSize, nil
}

// CheckSubcomponents verifies all subcomponents
func (mcc *managedConsensusComponents) CheckSubcomponents() error {
	mcc.mutConsensusComponents.Lock()
	defer mcc.mutConsensusComponents.Unlock()

	if mcc.consensusComponents == nil {
		return errors.ErrNilConsensusComponentsHolder
	}
	if check.IfNil(mcc.chronology) {
		return errors.ErrNilChronologyHandler
	}
	if check.IfNil(mcc.worker) {
		return errors.ErrNilConsensusWorker
	}
	if check.IfNil(mcc.broadcastMessenger) {
		return errors.ErrNilBroadcastMessenger
	}

	return nil
}

// HardforkTrigger returns the hardfork trigger
func (mcc *managedConsensusComponents) HardforkTrigger() HardforkTrigger {
	mcc.mutConsensusComponents.RLock()
	defer mcc.mutConsensusComponents.RUnlock()

	if mcc.consensusComponents == nil {
		return nil
	}

	return mcc.consensusComponents.hardforkTrigger
}

// IsInterfaceNil returns true if the underlying object is nil
func (mcc *managedConsensusComponents) IsInterfaceNil() bool {
	return mcc == nil
}

// String returns the name of the component
func (mcc *managedConsensusComponents) String() string {
	return "managedConsensusComponents"
}
