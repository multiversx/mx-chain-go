package factory

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var _ ComponentHandler = (*managedBootstrapComponents)(nil)
var _ BootstrapComponentsHolder = (*managedBootstrapComponents)(nil)
var _ BootstrapComponentsHandler = (*managedBootstrapComponents)(nil)

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

	if mbf.bootstrapComponents != nil {
		err := mbf.bootstrapComponents.Close()
		if err != nil {
			return err
		}
		mbf.bootstrapComponents = nil
	}

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mbf *managedBootstrapComponents) CheckSubcomponents() error {
	mbf.mutBootstrapComponents.Lock()
	defer mbf.mutBootstrapComponents.Unlock()

	if mbf.bootstrapComponents == nil {
		return errors.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(mbf.epochStartBootstraper) {
		return errors.ErrNilEpochStartBootstrapper
	}
	if check.IfNil(mbf.bootstrapParamsHandler) {
		return errors.ErrNilBootstrapParamsHandler
	}

	return nil
}

// EpochStartBootstrapper returns the epoch start bootstrapper
func (mbf *managedBootstrapComponents) EpochStartBootstrapper() EpochStartBootstrapper {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	return mbf.bootstrapComponents.epochStartBootstraper
}

// EpochBootstrapParams returns the epoch start bootstrap parameters handler
func (mbf *managedBootstrapComponents) EpochBootstrapParams() BootstrapParamsHandler {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return nil
	}

	return mbf.bootstrapComponents.bootstrapParamsHandler
}

// IsInterfaceNil returns true if the underlying object is nil
func (mbf *managedBootstrapComponents) IsInterfaceNil() bool {
	return mbf == nil
}
