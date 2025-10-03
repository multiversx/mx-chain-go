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

// Close will close all underlying subcomponents
func (m *managedProcessComponents) Close() error {
	m.mutProcessComponents.Lock()
	defer m.mutProcessComponents.Unlock()

	if m.processComponents == nil {
		return nil
	}

	err := m.Close()
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
	if check.IfNil(m.nodesCoordinator) {
		return errors.ErrNilNodesCoordinator
	}
	if check.IfNil(m.shardCoordinator) {
		return errors.ErrNilShardCoordinator
	}
	if check.IfNil(m.mainInterceptorsContainer) {
		return fmt.Errorf("%w on main network", errors.ErrNilInterceptorsContainer)
	}
	if check.IfNil(m.fullArchiveInterceptorsContainer) {
		return fmt.Errorf("%w on full archive network", errors.ErrNilInterceptorsContainer)
	}
	if check.IfNil(m.resolversContainer) {
		return errors.ErrNilResolversContainer
	}
	if check.IfNil(m.requestersFinder) {
		return errors.ErrNilRequestersFinder
	}
	if check.IfNil(m.roundHandler) {
		return errors.ErrNilRoundHandler
	}
	if check.IfNil(m.epochStartTrigger) {
		return errors.ErrNilEpochStartTrigger
	}
	if check.IfNil(m.epochStartNotifier) {
		return errors.ErrNilEpochStartNotifier
	}
	if check.IfNil(m.forkDetector) {
		return errors.ErrNilForkDetector
	}
	if check.IfNil(m.blockProcessor) {
		return errors.ErrNilBlockProcessor
	}
	if check.IfNil(m.blackListHandler) {
		return errors.ErrNilBlackListHandler
	}
	if check.IfNil(m.bootStorer) {
		return errors.ErrNilBootStorer
	}
	if check.IfNil(m.headerSigVerifier) {
		return errors.ErrNilHeaderSigVerifier
	}
	if check.IfNil(m.headerIntegrityVerifier) {
		return errors.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(m.validatorsStatistics) {
		return errors.ErrNilValidatorsStatistics
	}
	if check.IfNil(m.validatorsProvider) {
		return errors.ErrNilValidatorsProvider
	}
	if check.IfNil(m.blockTracker) {
		return errors.ErrNilBlockTracker
	}
	if check.IfNil(m.pendingMiniBlocksHandler) {
		return errors.ErrNilPendingMiniBlocksHandler
	}
	if check.IfNil(m.requestHandler) {
		return errors.ErrNilRequestHandler
	}
	if check.IfNil(m.txLogsProcessor) {
		return errors.ErrNilTxLogsProcessor
	}
	if check.IfNil(m.headerConstructionValidator) {
		return errors.ErrNilHeaderConstructionValidator
	}
	if check.IfNil(m.mainPeerShardMapper) {
		return fmt.Errorf("%w for main", errors.ErrNilPeerShardMapper)
	}
	if check.IfNil(m.fullArchivePeerShardMapper) {
		return fmt.Errorf("%w for full archive", errors.ErrNilPeerShardMapper)
	}
	if check.IfNil(m.fallbackHeaderValidator) {
		return errors.ErrNilFallbackHeaderValidator
	}
	if check.IfNil(m.nodeRedundancyHandler) {
		return errors.ErrNilNodeRedundancyHandler
	}
	if check.IfNil(m.currentEpochProvider) {
		return errors.ErrNilCurrentEpochProvider
	}
	if check.IfNil(m.scheduledTxsExecutionHandler) {
		return errors.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(m.txsSender) {
		return errors.ErrNilTxsSender
	}
	if check.IfNil(m.processedMiniBlocksTracker) {
		return process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(m.esdtDataStorageForApi) {
		return errors.ErrNilESDTDataStorage
	}
	if check.IfNil(m.sentSignaturesTracker) {
		return errors.ErrNilSentSignatureTracker
	}
	if check.IfNil(m.epochSystemSCProcessor) {
		return errors.ErrNilEpochSystemSCProcessor
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

	return m.nodesCoordinator
}

// ShardCoordinator returns the shard coordinator
func (m *managedProcessComponents) ShardCoordinator() sharding.Coordinator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.shardCoordinator
}

// InterceptorsContainer returns the interceptors container on the main network
func (m *managedProcessComponents) InterceptorsContainer() process.InterceptorsContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.mainInterceptorsContainer
}

// FullArchiveInterceptorsContainer returns the interceptors container on the full archive network
func (m *managedProcessComponents) FullArchiveInterceptorsContainer() process.InterceptorsContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.fullArchiveInterceptorsContainer
}

// ResolversContainer returns the resolvers container
func (m *managedProcessComponents) ResolversContainer() dataRetriever.ResolversContainer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.resolversContainer
}

// RequestersFinder returns the requesters finder
func (m *managedProcessComponents) RequestersFinder() dataRetriever.RequestersFinder {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.requestersFinder
}

// RoundHandler returns the roundHandler
func (m *managedProcessComponents) RoundHandler() consensus.RoundHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.roundHandler
}

// EpochStartTrigger returns the epoch start trigger handler
func (m *managedProcessComponents) EpochStartTrigger() epochStart.TriggerHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.epochStartTrigger
}

// EpochStartNotifier returns the epoch start notifier
func (m *managedProcessComponents) EpochStartNotifier() factory.EpochStartNotifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.epochStartNotifier
}

// ForkDetector returns the fork detector
func (m *managedProcessComponents) ForkDetector() process.ForkDetector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.forkDetector
}

// BlockProcessor returns the block processor
func (m *managedProcessComponents) BlockProcessor() process.BlockProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.blockProcessor
}

// BlockchainHook returns the block chain hook
func (m *managedProcessComponents) BlockchainHook() process.BlockChainHookWithAccountsAdapter {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	if check.IfNil(m.vmFactoryForProcessing) {
		return nil
	}

	return m.vmFactoryForProcessing.BlockChainHookImpl()
}

// BlackListHandler returns the black list handler
func (m *managedProcessComponents) BlackListHandler() process.TimeCacher {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.blackListHandler
}

// BootStorer returns the boot storer
func (m *managedProcessComponents) BootStorer() process.BootStorer {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.bootStorer
}

// HeaderSigVerifier returns the header signature verification
func (m *managedProcessComponents) HeaderSigVerifier() process.InterceptedHeaderSigVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.headerSigVerifier
}

// HeaderIntegrityVerifier returns the header integrity verifier
func (m *managedProcessComponents) HeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.headerIntegrityVerifier
}

// ValidatorsStatistics returns the validator statistics processor
func (m *managedProcessComponents) ValidatorsStatistics() process.ValidatorStatisticsProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.validatorsStatistics
}

// ValidatorsProvider returns the validator provider
func (m *managedProcessComponents) ValidatorsProvider() process.ValidatorsProvider {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.validatorsProvider
}

// BlockTracker returns the block tracker
func (m *managedProcessComponents) BlockTracker() process.BlockTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.blockTracker
}

// PendingMiniBlocksHandler returns the pending mini blocks handler
func (m *managedProcessComponents) PendingMiniBlocksHandler() process.PendingMiniBlocksHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.pendingMiniBlocksHandler
}

// RequestHandler returns the request handler
func (m *managedProcessComponents) RequestHandler() process.RequestHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.requestHandler
}

// TxLogsProcessor returns the tx logs processor
func (m *managedProcessComponents) TxLogsProcessor() process.TransactionLogProcessorDatabase {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.txLogsProcessor
}

// HeaderConstructionValidator returns the validator for header construction
func (m *managedProcessComponents) HeaderConstructionValidator() process.HeaderConstructionValidator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.headerConstructionValidator
}

// PeerShardMapper returns the peer to shard mapper of the main network
func (m *managedProcessComponents) PeerShardMapper() process.NetworkShardingCollector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.mainPeerShardMapper
}

// FullArchivePeerShardMapper returns the peer to shard mapper of the full archive network
func (m *managedProcessComponents) FullArchivePeerShardMapper() process.NetworkShardingCollector {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.fullArchivePeerShardMapper
}

// FallbackHeaderValidator returns the fallback header validator
func (m *managedProcessComponents) FallbackHeaderValidator() process.FallbackHeaderValidator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.fallbackHeaderValidator
}

// APITransactionEvaluator returns the api transaction evaluator
func (m *managedProcessComponents) APITransactionEvaluator() factory.TransactionEvaluator {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.apiTransactionEvaluator
}

// WhiteListHandler returns the white list handler
func (m *managedProcessComponents) WhiteListHandler() process.WhiteListHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.whiteListHandler
}

// WhiteListerVerifiedTxs returns the white lister verified txs
func (m *managedProcessComponents) WhiteListerVerifiedTxs() process.WhiteListHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.whiteListerVerifiedTxs
}

// HistoryRepository returns the history repository
func (m *managedProcessComponents) HistoryRepository() dblookupext.HistoryRepository {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.historyRepository
}

// ImportStartHandler returns the import status handler
func (m *managedProcessComponents) ImportStartHandler() update.ImportStartHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.importStartHandler
}

// RequestedItemsHandler returns the items handler for the requests
func (m *managedProcessComponents) RequestedItemsHandler() dataRetriever.RequestedItemsHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.requestedItemsHandler
}

// NodeRedundancyHandler returns the node redundancy handler
func (m *managedProcessComponents) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.nodeRedundancyHandler
}

// AccountsParser returns the genesis accounts parser
func (m *managedProcessComponents) AccountsParser() genesis.AccountsParser {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.accountsParser
}

// CurrentEpochProvider returns the current epoch provider that can decide if an epoch is active or not on the network
func (m *managedProcessComponents) CurrentEpochProvider() process.CurrentNetworkEpochProviderHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.currentEpochProvider
}

// ScheduledTxsExecutionHandler returns the scheduled transactions execution handler
func (m *managedProcessComponents) ScheduledTxsExecutionHandler() process.ScheduledTxsExecutionHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.scheduledTxsExecutionHandler
}

// TxsSenderHandler returns the transactions sender handler
func (m *managedProcessComponents) TxsSenderHandler() process.TxsSenderHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.txsSender
}

// HardforkTrigger returns the hardfork trigger
func (m *managedProcessComponents) HardforkTrigger() factory.HardforkTrigger {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.hardforkTrigger
}

// ProcessedMiniBlocksTracker returns the processed mini blocks tracker
func (m *managedProcessComponents) ProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.processedMiniBlocksTracker
}

// ESDTDataStorageHandlerForAPI returns the esdt data storage handler to be used for API calls
func (m *managedProcessComponents) ESDTDataStorageHandlerForAPI() vmcommon.ESDTNFTStorageHandler {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.esdtDataStorageForApi
}

// ReceiptsRepository returns the receipts repository
func (m *managedProcessComponents) ReceiptsRepository() factory.ReceiptsRepository {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.receiptsRepository
}

// SentSignaturesTracker returns the signature tracker
func (m *managedProcessComponents) SentSignaturesTracker() process.SentSignaturesTracker {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.sentSignaturesTracker
}

// EpochSystemSCProcessor returns the epoch start system SC processor
func (m *managedProcessComponents) EpochSystemSCProcessor() process.EpochStartSystemSCProcessor {
	m.mutProcessComponents.RLock()
	defer m.mutProcessComponents.RUnlock()

	if m.processComponents == nil {
		return nil
	}

	return m.epochSystemSCProcessor
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedProcessComponents) IsInterfaceNil() bool {
	return m == nil
}

// String returns the name of the component
func (m *managedProcessComponents) String() string {
	return factory.ProcessComponentsName
}
