package bootstrap

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
)

var _ factory.ComponentHandler = (*managedBootstrapComponents)(nil)
var _ factory.BootstrapComponentsHolder = (*managedBootstrapComponents)(nil)
var _ factory.BootstrapComponentsHandler = (*managedBootstrapComponents)(nil)

type managedBootstrapComponents struct {
	*bootstrapComponents
	bootstrapComponentsFactory *bootstrapComponentsFactory
	mutBootstrapComponents     sync.RWMutex
}

// NewManagedBootstrapComponents creates a managed bootstrap components handler
func NewManagedBootstrapComponents(bootstrapComponentsFactory *bootstrapComponentsFactory) (*managedBootstrapComponents, error) {
	if bootstrapComponentsFactory == nil {
		return nil, errors.ErrNilBootstrapComponentsFactory
	}

	return &managedBootstrapComponents{
		bootstrapComponents:        nil,
		bootstrapComponentsFactory: bootstrapComponentsFactory,
	}, nil
}

// Create creates the bootstrap components
func (mbf *managedBootstrapComponents) Create() error {
	bc, err := mbf.bootstrapComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrBootstrapDataComponentsFactoryCreate, err)
	}

	mbf.mutBootstrapComponents.Lock()
	mbf.bootstrapComponents = bc
	mbf.mutBootstrapComponents.Unlock()

	return nil
}

// Close closes all the consensus components
func (mbf *managedBootstrapComponents) Close() error {
	mbf.mutBootstrapComponents.Lock()
	defer mbf.mutBootstrapComponents.Unlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	err := mbf.bootstrapComponents.Close()
	if err != nil {
		return err
	}
	mbf.bootstrapComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mbf *managedBootstrapComponents) CheckSubcomponents() error {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return errors.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(mbf.epochStartBootstrapper) {
		return errors.ErrNilEpochStartBootstrapper
	}
	if check.IfNil(mbf.bootstrapParamsHolder) {
		return errors.ErrNilBootstrapParamsHandler
	}

	return nil
}

// EpochStartBootstrapper returns the epoch start bootstrapper
func (mbf *managedBootstrapComponents) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	return mbf.bootstrapComponents.epochStartBootstrapper
}

// GuardedAccountHandler returns the guarded account handler
func (mbf *managedBootstrapComponents) GuardedAccountHandler() process.GuardedAccountHandler {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	return mbf.bootstrapComponents.guardedAccountHandler
}

// EpochBootstrapParams returns the epoch start bootstrap parameters handler
func (mbf *managedBootstrapComponents) EpochBootstrapParams() factory.BootstrapParamsHolder {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	return mbf.bootstrapComponents.bootstrapParamsHolder
}

// IsInterfaceNil returns true if the underlying object is nil
func (mbf *managedBootstrapComponents) IsInterfaceNil() bool {
	return mbf == nil
}

// String returns the name of the component
func (mbf *managedBootstrapComponents) String() string {
	return factory.BootstrapComponentsName
}
