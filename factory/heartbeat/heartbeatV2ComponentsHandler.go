package heartbeat

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
)

type managedHeartbeatV2Components struct {
	*heartbeatV2Components
	heartbeatV2ComponentsFactory *heartbeatV2ComponentsFactory
	mutHeartbeatV2Components     sync.RWMutex
}

// NewManagedHeartbeatV2Components creates a new heartbeatV2 components handler
func NewManagedHeartbeatV2Components(hcf *heartbeatV2ComponentsFactory) (*managedHeartbeatV2Components, error) {
	if hcf == nil {
		return nil, errors.ErrNilHeartbeatV2ComponentsFactory
	}

	return &managedHeartbeatV2Components{
		heartbeatV2Components:        nil,
		heartbeatV2ComponentsFactory: hcf,
	}, nil
}

// Create creates the heartbeatV2 components
func (mhc *managedHeartbeatV2Components) Create() error {
	hc, err := mhc.heartbeatV2ComponentsFactory.Create()
	if err != nil {
		return err
	}

	mhc.mutHeartbeatV2Components.Lock()
	mhc.heartbeatV2Components = hc
	mhc.mutHeartbeatV2Components.Unlock()

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mhc *managedHeartbeatV2Components) CheckSubcomponents() error {
	mhc.mutHeartbeatV2Components.RLock()
	defer mhc.mutHeartbeatV2Components.RUnlock()

	if mhc.heartbeatV2Components == nil {
		return errors.ErrNilHeartbeatV2Components
	}
	if check.IfNil(mhc.sender) {
		return errors.ErrNilHeartbeatV2Sender
	}

	return nil
}

// String returns the name of the component
func (mhc *managedHeartbeatV2Components) String() string {
	return factory.HeartbeatV2ComponentsName
}

// Monitor returns the heartbeatV2 monitor
func (mhc *managedHeartbeatV2Components) Monitor() factory.HeartbeatV2Monitor {
	mhc.mutHeartbeatV2Components.Lock()
	defer mhc.mutHeartbeatV2Components.Unlock()

	if mhc.heartbeatV2Components == nil {
		return nil
	}

	return mhc.monitor
}

// Close closes the heartbeat components
func (mhc *managedHeartbeatV2Components) Close() error {
	mhc.mutHeartbeatV2Components.Lock()
	defer mhc.mutHeartbeatV2Components.Unlock()

	if mhc.heartbeatV2Components == nil {
		return nil
	}

	err := mhc.heartbeatV2Components.Close()
	if err != nil {
		return err
	}
	mhc.heartbeatV2Components = nil

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mhc *managedHeartbeatV2Components) IsInterfaceNil() bool {
	return mhc == nil
}
