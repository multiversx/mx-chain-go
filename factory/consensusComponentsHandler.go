package factory

import (
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
func (mcf *managedConsensusComponents) Create() error {
	cc, err := mcf.consensusComponentsFactory.Create()
	if err != nil {
		return err
	}

	mcf.mutConsensusComponents.Lock()
	mcf.consensusComponents = cc
	mcf.mutConsensusComponents.Unlock()

	return nil
}

// Close closes all the consensus components
func (mcf *managedConsensusComponents) Close() error {
	mcf.mutConsensusComponents.Lock()
	defer mcf.mutConsensusComponents.Unlock()

	if mcf.consensusComponents == nil {
		err := mcf.consensusComponents.Close()
		if err != nil {
			return err
		}
		mcf.consensusComponents = nil
	}

	return nil
}

// Chronology returns the chronology handler
func (mcf *managedConsensusComponents) Chronology() consensus.ChronologyHandler {
	mcf.mutConsensusComponents.RLock()
	defer mcf.mutConsensusComponents.Unlock()

	if mcf.consensusComponents == nil {
		return nil
	}

	return mcf.consensusComponents.chronology
}

// ConsensusWorker returns the consensus worker
func (mcf *managedConsensusComponents) ConsensusWorker() ConsensusWorker {
	mcf.mutConsensusComponents.RLock()
	defer mcf.mutConsensusComponents.Unlock()

	if mcf.consensusComponents == nil {
		return nil
	}

	return mcf.consensusComponents.worker
}

// BroadcastMessenger returns the consensus broadcast messenger
func (mcf *managedConsensusComponents) BroadcastMessenger() consensus.BroadcastMessenger {
	mcf.mutConsensusComponents.RLock()
	defer mcf.mutConsensusComponents.Unlock()

	if mcf.consensusComponents == nil {
		return nil
	}

	return mcf.consensusComponents.broadcastMessenger
}

// CheckSubcomponents verifies all subcomponents
func (mcf *managedConsensusComponents) CheckSubcomponents() error {
	mcf.mutConsensusComponents.Lock()
	defer mcf.mutConsensusComponents.Unlock()

	if mcf.consensusComponents == nil {
		return errors.ErrNilConsensusComponentsHolder
	}
	if check.IfNil(mcf.chronology) {
		return errors.ErrNilChronologyHandler
	}
	if check.IfNil(mcf.worker) {
		return errors.ErrNilConsensusWorker
	}
	if check.IfNil(mcf.broadcastMessenger) {
		return errors.ErrNilBroadcastMessenger
	}

	return nil
}

// IsInterfaceNil returns true if the underlying object is nil
func (mcf *managedConsensusComponents) IsInterfaceNil() bool {
	return mcf == nil
}
