package factory

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ ComponentHandler = (*managedProcessComponents)(nil)
var _ ProcessComponentsHolder = (*managedProcessComponents)(nil)
var _ ProcessComponentsHandler = (*managedProcessComponents)(nil)

type managedProcessComponents struct {
	*processComponents
	factory              *processComponentsFactory
	mutProcessComponents sync.RWMutex
}

// NewManagedProcessComponents returns a news instance of managedProcessComponents
func NewManagedProcessComponents(pcf *processComponentsFactory) (*managedProcessComponents, error) {
	if pcf == nil {
		return nil, errors.ErrNilProcessComponentsFactory
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

	if m.processComponents == nil {
		return nil
	}

	err := m.processComponents.Close()
	if err != nil {
		return err
	}
	m.processComponents = nil

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
	if check.IfNil(m.processComponents.roundHandler) {
		return errors.ErrNilRoundHandler
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
	if check.IfNil(m.processComponents.fallbackHeaderValidator) {
		return errors.ErrNilFallbackHeaderValidator
	}
	if check.IfNil(m.processComponents.nodeRedundancyHandler) {
		return errors.ErrNilNodeRedundancyHandler
	}
	if check.IfNil(m.processComponents.currentEpochProvider) {
		return errors.ErrNilCurrentEpochProvider
	}
	if check.IfNil(m.processComponents.scheduledTxsExecutionHandler) {
		return errors.ErrNilScheduledTxsExecutionHandler
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

// RoundHandler returns the roundHandler
func (m *managedProcessComponents) RoundHandler() consensus.RoundHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.roundHandler
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

// FallbackHeaderValidator returns the fallback header validator
func (m *managedProcessComponents) FallbackHeaderValidator() process.FallbackHeaderValidator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.fallbackHeaderValidator
}

// TransactionSimulatorProcessor returns the transaction simulator processor
func (m *managedProcessComponents) TransactionSimulatorProcessor() TransactionSimulatorProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.txSimulatorProcessor
}

// WhiteListHandler returns the white list handler
func (m *managedProcessComponents) WhiteListHandler() process.WhiteListHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.whiteListHandler
}

// WhiteListerVerifiedTxs returns the white lister verified txs
func (m *managedProcessComponents) WhiteListerVerifiedTxs() process.WhiteListHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.whiteListerVerifiedTxs
}

// HistoryRepository returns the history repository
func (m *managedProcessComponents) HistoryRepository() dblookupext.HistoryRepository {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.historyRepository
}

// ImportStartHandler returns the import status handler
func (m *managedProcessComponents) ImportStartHandler() update.ImportStartHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.importStartHandler
}

// RequestedItemsHandler returns the items handler for the requests
func (m *managedProcessComponents) RequestedItemsHandler() dataRetriever.RequestedItemsHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.requestedItemsHandler
}

// NodeRedundancyHandler returns the node redundancy handler
func (m *managedProcessComponents) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.nodeRedundancyHandler
}

// CurrentEpochProvider returns the current epoch provider that can decide if an epoch is active or not on the network
func (m *managedProcessComponents) CurrentEpochProvider() process.CurrentNetworkEpochProviderHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.currentEpochProvider
}

// ScheduledTxsExecutionHandler returns the scheduled transactions execution handler
func (m *managedProcessComponents) ScheduledTxsExecutionHandler() process.ScheduledTxsExecutionHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.scheduledTxsExecutionHandler
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedProcessComponents) IsInterfaceNil() bool {
	return m == nil
}

// String returns the name of the component
func (m *managedProcessComponents) String() string {
	return "managedProcessComponents"
}
