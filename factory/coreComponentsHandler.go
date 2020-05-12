package factory

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// CoreComponentsHandlerArgs holds the arguments required to create a core components handler
type CoreComponentsHandlerArgs CoreComponentsFactoryArgs

// coreComponents is the DTO used for core components
type coreComponents struct {
	Hasher                   hashing.Hasher
	InternalMarshalizer      marshal.Marshalizer
	VmMarshalizer            marshal.Marshalizer
	TxSignMarshalizer        marshal.Marshalizer
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}

// ManagedCoreComponents is an implementation of core components handler that can create, close and access the core components
type ManagedCoreComponents struct {
	coreComponentsFactory *CoreComponentsFactory
	*coreComponents
	cancelFunc        func()
	mutCoreComponents sync.RWMutex
}

// NewManagedCoreComponents creates a new core components handler implementation
func NewManagedCoreComponents(args CoreComponentsHandlerArgs) (*ManagedCoreComponents, error) {
	coreComponentsFactory := NewCoreComponentsFactory(CoreComponentsFactoryArgs(args))
	mcc := &ManagedCoreComponents{
		coreComponents:        nil,
		coreComponentsFactory: coreComponentsFactory,
	}
	return mcc, nil
}

// Create creates the core components
func (mcc *ManagedCoreComponents) Create() error {
	cc, err := mcc.coreComponentsFactory.Create()
	if err != nil {
		return err
	}

	mcc.mutCoreComponents.Lock()
	mcc.coreComponents = cc
	_, mcc.cancelFunc = context.WithCancel(context.Background())
	mcc.mutCoreComponents.Unlock()

	return nil
}

// Close closes the managed core components
func (mcc *ManagedCoreComponents) Close() error {
	mcc.mutCoreComponents.Lock()
	mcc.coreComponents.StatusHandler.Close()
	mcc.cancelFunc()
	mcc.cancelFunc = nil
	mcc.coreComponents = nil
	mcc.mutCoreComponents.Unlock()

	return nil
}

// InternalMarshalizer returns the core components internal marshalizer
func (mcc *ManagedCoreComponents) InternalMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.InternalMarshalizer
}

// TxMarshalizer returns the core components tx marshalizer
func (mcc *ManagedCoreComponents) TxMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.TxSignMarshalizer
}

// VmMarshalizer returns the core components vm marshalizer
func (mcc *ManagedCoreComponents) VmMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.VmMarshalizer
}

// Hasher returns the core components Hasher
func (mcc *ManagedCoreComponents) Hasher() hashing.Hasher {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.Hasher
}

// Uint64ByteSliceConverter returns the core component converter between a byte slice and uint64
func (mcc *ManagedCoreComponents) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.Uint64ByteSliceConverter
}

// StatusHandler returns the core components status handler
func (mcc *ManagedCoreComponents) StatusHandler() core.AppStatusHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.StatusHandler
}

// SetStatusHandler allows the change of the status handler
func (mcc *ManagedCoreComponents) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if check.IfNil(statusHandler) {
		return ErrNilStatusHandler
	}

	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents == nil {
		return ErrNilCoreComponents
	}

	mcc.coreComponents.StatusHandler = statusHandler

	return nil
}

// ChainID returns the core components chainID
func (mcc *ManagedCoreComponents) ChainID() []byte {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.ChainID
}
