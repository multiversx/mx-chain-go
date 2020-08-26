package factory

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ ComponentHandler = (*managedProcessComponents)(nil)
var _ ProcessComponentsHolder = (*managedProcessComponents)(nil)
var _ ProcessComponentsHandler = (*managedProcessComponents)(nil)

// TODO: integrate this in main.go and remove obsolete component from structs.go afterwards

type managedProcessComponents struct {
	*processComponents
	factory              *processComponentsFactory
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
	m.mutProcessComponents.Unlock()

	return nil
}

// Close will close all underlying sub-components
func (m *managedProcessComponents) Close() error {
	m.mutProcessComponents.Lock()
	defer m.mutProcessComponents.Unlock()

	if m.processComponents != nil {
		err := m.processComponents.Close()
		if err != nil {
			return err
		}
		m.processComponents = nil
	}

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (m *managedProcessComponents) CheckSubcomponents() error {
	m.mutProcessComponents.Lock()
	defer m.mutProcessComponents.Unlock()

	if m.processComponents == nil {
		return errors.ErrNilProcessComponents
	}
	if check.IfNil(m.processComponents.nodesCoordinator) {
		return errors.ErrNilNodesCoordinator
	}
	if check.IfNil(m.processComponents.shardCoordinator) {
		return errors.ErrNilShardCoordinator
	}
	if check.IfNil(m.processComponents.interceptorsContainer) {
		return errors.ErrNilInterceptorsContainer
	}
	if check.IfNil(m.processComponents.resolversFinder) {
		return errors.ErrNilResolversFinder
	}
	if check.IfNil(m.processComponents.rounder) {
		return errors.ErrNilRounder
	}
	if check.IfNil(m.processComponents.epochStartTrigger) {
		return errors.ErrNilEpochStartTrigger
	}
	if check.IfNil(m.processComponents.epochStartNotifier) {
		return errors.ErrNilEpochStartNotifier
	}
	if check.IfNil(m.processComponents.forkDetector) {
		return errors.ErrNilForkDetector
	}
	if check.IfNil(m.processComponents.blockProcessor) {
		return errors.ErrNilBlockProcessor
	}
	if check.IfNil(m.processComponents.blackListHandler) {
		return errors.ErrNilBlackListHandler
	}
	if check.IfNil(m.processComponents.bootStorer) {
		return errors.ErrNilBootStorer
	}
	if check.IfNil(m.processComponents.headerSigVerifier) {
		return errors.ErrNilHeaderSigVerifier
	}
	if check.IfNil(m.processComponents.headerIntegrityVerifier) {
		return errors.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(m.processComponents.validatorsStatistics) {
		return errors.ErrNilValidatorsStatistics
	}
	if check.IfNil(m.processComponents.validatorsProvider) {
		return errors.ErrNilValidatorsProvider
	}
	if check.IfNil(m.processComponents.blockTracker) {
		return errors.ErrNilBlockTracker
	}
	if check.IfNil(m.processComponents.pendingMiniBlocksHandler) {
		return errors.ErrNilPendingMiniBlocksHandler
	}
	if check.IfNil(m.processComponents.requestHandler) {
		return errors.ErrNilRequestHandler
	}
	if check.IfNil(m.processComponents.txLogsProcessor) {
		return errors.ErrNilTxLogsProcessor
	}
	if check.IfNil(m.processComponents.headerConstructionValidator) {
		return errors.ErrNilHeaderConstructionValidator
	}
	if check.IfNil(m.processComponents.peerShardMapper) {
		return errors.ErrNilPeerShardMapper
	}

	return nil
}

// NodesCoordinator returns the nodes coordinator
func (m *managedProcessComponents) NodesCoordinator() sharding.NodesCoordinator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.nodesCoordinator
}

// ShardCoordinator returns the shard coordinator
func (m *managedProcessComponents) ShardCoordinator() sharding.Coordinator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.shardCoordinator
}

// InterceptorsContainer returns the interceptors container
func (m *managedProcessComponents) InterceptorsContainer() process.InterceptorsContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.interceptorsContainer
}

// ResolversFinder returns the resolvers finder
func (m *managedProcessComponents) ResolversFinder() dataRetriever.ResolversFinder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.resolversFinder
}

// Rounder returns the rounderer
func (m *managedProcessComponents) Rounder() consensus.Rounder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.rounder
}

// EpochStartTrigger returns the epoch start trigger handler
func (m *managedProcessComponents) EpochStartTrigger() epochStart.TriggerHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.epochStartTrigger
}

// EpochStartNotifier returns the epoch start notifier
func (m *managedProcessComponents) EpochStartNotifier() EpochStartNotifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.epochStartNotifier
}

// ForkDetector returns the fork detector
func (m *managedProcessComponents) ForkDetector() process.ForkDetector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.forkDetector
}

// BlockProcessor returns the block processor
func (m *managedProcessComponents) BlockProcessor() process.BlockProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.blockProcessor
}

// BlackListHandler returns the black list handler
func (m *managedProcessComponents) BlackListHandler() process.TimeCacher {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.blackListHandler
}

// BootStorer returns the boot storer
func (m *managedProcessComponents) BootStorer() process.BootStorer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.bootStorer
}

// HeaderSigVerifier returns the header signature verification
func (m *managedProcessComponents) HeaderSigVerifier() process.InterceptedHeaderSigVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.headerSigVerifier
}

// HeaderIntegrityVerifier returns the header integrity verifier
func (m *managedProcessComponents) HeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.headerIntegrityVerifier
}

// ValidatorsStatistics returns the validator statistics processor
func (m *managedProcessComponents) ValidatorsStatistics() process.ValidatorStatisticsProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.validatorsStatistics
}

// ValidatorsProvider returns the validator provider
func (m *managedProcessComponents) ValidatorsProvider() process.ValidatorsProvider {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.validatorsProvider
}

// BlockTracker returns the block tracker
func (m *managedProcessComponents) BlockTracker() process.BlockTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.blockTracker
}

// PendingMiniBlocksHandler returns the pending mini blocks handler
func (m *managedProcessComponents) PendingMiniBlocksHandler() process.PendingMiniBlocksHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.pendingMiniBlocksHandler
}

// RequestHandler returns the request handler
func (m *managedProcessComponents) RequestHandler() process.RequestHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.requestHandler
}

// TxLogsProcessor returns the tx logs processor
func (m *managedProcessComponents) TxLogsProcessor() process.TransactionLogProcessorDatabase {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.txLogsProcessor
}

// HeaderConstructionValidator returns the validator for header construction
func (m *managedProcessComponents) HeaderConstructionValidator() process.HeaderConstructionValidator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.headerConstructionValidator
}

// PeerShardMapper returns the peer to shard mapper
func (m *managedProcessComponents) PeerShardMapper() process.NetworkShardingCollector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.peerShardMapper
}

func (m *managedProcessComponents) TransactionSimulatorProcessor() TransactionSimulatorProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.txSimulatorProcessor
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedProcessComponents) IsInterfaceNil() bool {
	return m == nil
}
