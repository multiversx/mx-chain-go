package processing

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/update"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ factory.ComponentHandler = (*managedProcessComponents)(nil)
var _ factory.ProcessComponentsHolder = (*managedProcessComponents)(nil)
var _ factory.ProcessComponentsHandler = (*managedProcessComponents)(nil)

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
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

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
	if check.IfNil(m.processComponents.resolversContainer) {
		return errors.ErrNilResolversContainer
	}
	if check.IfNil(m.processComponents.requestersFinder) {
		return errors.ErrNilRequestersFinder
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
	if check.IfNil(m.processComponents.mainPeerShardMapper) {
		return fmt.Errorf("%w for main", errors.ErrNilPeerShardMapper)
	}
	if check.IfNil(m.processComponents.fullArchivePeerShardMapper) {
		return fmt.Errorf("%w for full archive", errors.ErrNilPeerShardMapper)
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
	if check.IfNil(m.processComponents.txsSender) {
		return errors.ErrNilTxsSender
	}
	if check.IfNil(m.processComponents.processedMiniBlocksTracker) {
		return process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(m.processComponents.esdtDataStorageForApi) {
		return errors.ErrNilESDTDataStorage
	}

	return nil
}

// NodesCoordinator returns the nodes coordinator
func (m *managedProcessComponents) NodesCoordinator() nodesCoordinator.NodesCoordinator {
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

// ResolversContainer returns the resolvers container
func (m *managedProcessComponents) ResolversContainer() dataRetriever.ResolversContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.resolversContainer
}

// RequestersFinder returns the requesters finder
func (m *managedProcessComponents) RequestersFinder() dataRetriever.RequestersFinder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.requestersFinder
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
func (m *managedProcessComponents) EpochStartNotifier() factory.EpochStartNotifier {
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

// PeerShardMapper returns the peer to shard mapper of the main network
func (m *managedProcessComponents) PeerShardMapper() process.NetworkShardingCollector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.mainPeerShardMapper
}

// FullArchivePeerShardMapper returns the peer to shard mapper of the full archive network
func (m *managedProcessComponents) FullArchivePeerShardMapper() process.NetworkShardingCollector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.fullArchivePeerShardMapper
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
func (m *managedProcessComponents) TransactionSimulatorProcessor() factory.TransactionSimulatorProcessor {
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

// AccountsParser returns the genesis accounts parser
func (m *managedProcessComponents) AccountsParser() genesis.AccountsParser {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.accountsParser
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

// TxsSenderHandler returns the transactions sender handler
func (m *managedProcessComponents) TxsSenderHandler() process.TxsSenderHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.txsSender
}

// HardforkTrigger returns the hardfork trigger
func (m *managedProcessComponents) HardforkTrigger() factory.HardforkTrigger {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.hardforkTrigger
}

// ProcessedMiniBlocksTracker returns the processed mini blocks tracker
func (m *managedProcessComponents) ProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.processedMiniBlocksTracker
}

// ESDTDataStorageHandlerForAPI returns the esdt data storage handler to be used for API calls
func (m *managedProcessComponents) ESDTDataStorageHandlerForAPI() vmcommon.ESDTNFTStorageHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.esdtDataStorageForApi
}

// ReceiptsRepository returns the receipts repository
func (m *managedProcessComponents) ReceiptsRepository() factory.ReceiptsRepository {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processComponents.receiptsRepository
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedProcessComponents) IsInterfaceNil() bool {
	return m == nil
}

// String returns the name of the component
func (m *managedProcessComponents) String() string {
	return factory.ProcessComponentsName
}
