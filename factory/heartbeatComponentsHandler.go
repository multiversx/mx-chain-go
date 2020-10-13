package factory

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

var _ ComponentHandler = (*managedHeartbeatComponents)(nil)
var _ HeartbeatComponentsHolder = (*managedHeartbeatComponents)(nil)
var _ HeartbeatComponentsHandler = (*managedHeartbeatComponents)(nil)

type managedHeartbeatComponents struct {
	*heartbeatComponents
	heartbeatComponentsFactory *heartbeatComponentsFactory
	mutHeartbeatComponents     sync.RWMutex
}

// NewManagedHeartbeatComponents creates a new heartbeat components handler
func NewManagedHeartbeatComponents(hcf *heartbeatComponentsFactory) (*managedHeartbeatComponents, error) {
	if hcf == nil {
		return nil, errors.ErrNilHeartbeatComponentsFactory
	}

	return &managedHeartbeatComponents{
		heartbeatComponents:        nil,
		heartbeatComponentsFactory: hcf,
	}, nil
}

// Create creates the heartbeat components
func (mhc *managedHeartbeatComponents) Create() error {
	hc, err := mhc.heartbeatComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrHeartbeatComponentsFactoryCreate, err)
	}

	mhc.mutHeartbeatComponents.Lock()
	mhc.heartbeatComponents = hc
	mhc.mutHeartbeatComponents.Unlock()

	return nil
}

// Close closes the heartbeat components
func (mhc *managedHeartbeatComponents) Close() error {
	mhc.mutHeartbeatComponents.Lock()
	defer mhc.mutHeartbeatComponents.Unlock()

	if mhc.heartbeatComponents == nil {
		return nil
	}

	err := mhc.heartbeatComponents.Close()
	if err != nil {
		return err
	}
	mhc.heartbeatComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mhc *managedHeartbeatComponents) CheckSubcomponents() error {
	mhc.mutHeartbeatComponents.Lock()
	defer mhc.mutHeartbeatComponents.Unlock()

	if mhc.heartbeatComponents == nil {
		return errors.ErrNilHeartbeatComponents
	}
	if check.IfNil(mhc.messageHandler) {
		return errors.ErrNilHeartbeatMessageHandler
	}
	if check.IfNil(mhc.monitor) {
		return errors.ErrNilHeartbeatMonitor
	}
	if check.IfNil(mhc.sender) {
		return errors.ErrNilHeartbeatSender
	}
	if check.IfNil(mhc.storer) {
		return errors.ErrNilHeartbeatStorer
	}

	return nil
}

// MessageHandler returns the heartbeat message handler
func (mhc *managedHeartbeatComponents) MessageHandler() heartbeat.MessageHandler {
	mhc.mutHeartbeatComponents.RLock()
	defer mhc.mutHeartbeatComponents.RUnlock()

	if mhc.heartbeatComponents == nil {
		return nil
	}

	return mhc.messageHandler
}

// Monitor returns the heartbeat monitor
func (mhc *managedHeartbeatComponents) Monitor() HeartbeatMonitor {
	mhc.mutHeartbeatComponents.RLock()
	defer mhc.mutHeartbeatComponents.RUnlock()

	if mhc.heartbeatComponents == nil {
		return nil
	}

	return mhc.monitor
}

// Sender returns the heartbeat sender
func (mhc *managedHeartbeatComponents) Sender() HeartbeatSender {
	mhc.mutHeartbeatComponents.RLock()
	defer mhc.mutHeartbeatComponents.RUnlock()

	if mhc.heartbeatComponents == nil {
		return nil
	}

	return mhc.sender
}

// Storer returns the heartbeat storer
func (mhc *managedHeartbeatComponents) Storer() HeartbeatStorer {
	mhc.mutHeartbeatComponents.RLock()
	defer mhc.mutHeartbeatComponents.RUnlock()

	if mhc.heartbeatComponents == nil {
		return nil
	}

	return mhc.storer
}

// IsInterfaceNil returns true if the underlying object is nil
func (mhc *managedHeartbeatComponents) IsInterfaceNil() bool {
	return mhc == nil
}

// String returns the name of the component
func (mbf *managedHeartbeatComponents) String() string {
	return "managedHeartbeatComponents"
}
