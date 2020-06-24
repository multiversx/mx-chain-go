package factory

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TODO: integrate this in main.go and remove obsolete component from structs.go afterwards

type managedProcessComponents struct {
	*processComponents
	factory              *processComponentsFactory
	cancelFunc           func()
	mutProcessComponents sync.RWMutex
}

// NewManagedProcessComponents returns a news instance of managedProcessComponents
func NewManagedProcessComponents(args ProcessComponentsFactoryArgs) (*managedProcessComponents, error) {
	pcf, err := NewProcessComponentsFactory(args)
	if err != nil {
		return nil, err
	}

	return &managedProcessComponents{
		processComponents: nil,
		factory:           pcf,
	}, nil
}

// Create will create the managed components
func (m *managedProcessComponents) Create() error {
	pc, err := m.factory.Create()
	if err != nil {
		return err
	}

	m.mutProcessComponents.Lock()
	m.processComponents = pc
	_, m.cancelFunc = context.WithCancel(context.Background())
	m.mutProcessComponents.Unlock()

	return nil
}

// Close will close all underlying sub-components
func (m *managedProcessComponents) Close() error {
	m.mutProcessComponents.Lock()
	defer m.mutProcessComponents.Unlock()

	m.cancelFunc()
	//TODO: close underlying components
	m.cancelFunc = nil
	m.processComponents = nil

	return nil
}

// InterceptorsContainer returns the interceptors container
func (m *managedProcessComponents) InterceptorsContainer() process.InterceptorsContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.InterceptorsContainer
}

// ResolversFinder returns the resolvers finder
func (m *managedProcessComponents) ResolversFinder() dataRetriever.ResolversFinder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.ResolversFinder
}

// Rounder returns the rounderer
func (m *managedProcessComponents) Rounder() consensus.Rounder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.Rounder
}

// EpochStartTrigger returns the epoch start trigger
func (m *managedProcessComponents) EpochStartTrigger() epochStart.TriggerHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.EpochStartTrigger
}

// ForkDetector returns the fork detector
func (m *managedProcessComponents) ForkDetector() process.ForkDetector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.ForkDetector
}

// BlockProcessor returns the block processor
func (m *managedProcessComponents) BlockProcessor() process.BlockProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.BlockProcessor
}

// BlackListHandler returns the black list handler
func (m *managedProcessComponents) BlackListHandler() process.BlackListHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.BlackListHandler
}

// BootStorer returns the boot storer
func (m *managedProcessComponents) BootStorer() process.BootStorer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.BootStorer
}

// HeaderSigVerifier returns the header signature verification
func (m *managedProcessComponents) HeaderSigVerifier() process.InterceptedHeaderSigVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.HeaderSigVerifier
}

// HeaderIntegrityVerifier returns the header integrity verifier
func (m *managedProcessComponents) HeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.HeaderIntegrityVerifier
}

// ValidatorsStatistics returns the validator statistics processor
func (m *managedProcessComponents) ValidatorsStatistics() process.ValidatorStatisticsProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.ValidatorsStatistics
}

// ValidatorsProvider returns the validator provider
func (m *managedProcessComponents) ValidatorsProvider() process.ValidatorsProvider {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.ValidatorsProvider
}

// BlockTracker returns the block tracker
func (m *managedProcessComponents) BlockTracker() process.BlockTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.BlockTracker
}

// PendingMiniBlocksHandler returns the pending mini blocks handler
func (m *managedProcessComponents) PendingMiniBlocksHandler() process.PendingMiniBlocksHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.PendingMiniBlocksHandler
}

// RequestHandler returns the request handler
func (m *managedProcessComponents) RequestHandler() process.RequestHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.RequestHandler
}

// TxLogsProcessor returns the tx logs processor
func (m *managedProcessComponents) TxLogsProcessor() process.TransactionLogProcessorDatabase {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.TxLogsProcessor
}

func (m *managedProcessComponents) HeaderConstructionValidator() process.HeaderConstructionValidator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.HeaderConstructionValidator
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedProcessComponents) IsInterfaceNil() bool {
	return m == nil
}
