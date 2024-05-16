package mock

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/update"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ProcessComponentsMock -
type ProcessComponentsMock struct {
	NodesCoord                           nodesCoordinator.NodesCoordinator
	ShardCoord                           sharding.Coordinator
	IntContainer                         process.InterceptorsContainer
	FullArchiveIntContainer              process.InterceptorsContainer
	ResContainer                         dataRetriever.ResolversContainer
	ReqFinder                            dataRetriever.RequestersFinder
	RoundHandlerField                    consensus.RoundHandler
	EpochTrigger                         epochStart.TriggerHandler
	EpochNotifier                        factory.EpochStartNotifier
	ForkDetect                           process.ForkDetector
	BlockProcess                         process.BlockProcessor
	BlackListHdl                         process.TimeCacher
	BootSore                             process.BootStorer
	HeaderSigVerif                       process.InterceptedHeaderSigVerifier
	HeaderIntegrVerif                    process.HeaderIntegrityVerifier
	ValidatorStatistics                  process.ValidatorStatisticsProcessor
	ValidatorProvider                    process.ValidatorsProvider
	BlockTrack                           process.BlockTracker
	PendingMiniBlocksHdl                 process.PendingMiniBlocksHandler
	ReqHandler                           process.RequestHandler
	TxLogsProcess                        process.TransactionLogProcessorDatabase
	HeaderConstructValidator             process.HeaderConstructionValidator
	MainPeerMapper                       process.NetworkShardingCollector
	FullArchivePeerMapper                process.NetworkShardingCollector
	TransactionEvaluator                 factory.TransactionEvaluator
	FallbackHdrValidator                 process.FallbackHeaderValidator
	WhiteListHandlerInternal             process.WhiteListHandler
	WhiteListerVerifiedTxsInternal       process.WhiteListHandler
	HistoryRepositoryInternal            dblookupext.HistoryRepository
	ImportStartHandlerInternal           update.ImportStartHandler
	RequestedItemsHandlerInternal        dataRetriever.RequestedItemsHandler
	NodeRedundancyHandlerInternal        consensus.NodeRedundancyHandler
	CurrentEpochProviderInternal         process.CurrentNetworkEpochProviderHandler
	ScheduledTxsExecutionHandlerInternal process.ScheduledTxsExecutionHandler
	TxsSenderHandlerField                process.TxsSenderHandler
	HardforkTriggerField                 factory.HardforkTrigger
	ProcessedMiniBlocksTrackerInternal   process.ProcessedMiniBlocksTracker
	ESDTDataStorageHandlerForAPIInternal vmcommon.ESDTNFTStorageHandler
	AccountsParserInternal               genesis.AccountsParser
	ReceiptsRepositoryInternal           factory.ReceiptsRepository
	SentSignaturesTrackerInternal        process.SentSignaturesTracker
	EpochSystemSCProcessorInternal       process.EpochStartSystemSCProcessor
}

// Create -
func (pcm *ProcessComponentsMock) Create() error {
	return nil
}

// Close -
func (pcm *ProcessComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (pcm *ProcessComponentsMock) CheckSubcomponents() error {
	return nil
}

// NodesCoordinator -
func (pcm *ProcessComponentsMock) NodesCoordinator() nodesCoordinator.NodesCoordinator {
	return pcm.NodesCoord
}

// ShardCoordinator -
func (pcm *ProcessComponentsMock) ShardCoordinator() sharding.Coordinator {
	return pcm.ShardCoord
}

// InterceptorsContainer -
func (pcm *ProcessComponentsMock) InterceptorsContainer() process.InterceptorsContainer {
	return pcm.IntContainer
}

// FullArchiveInterceptorsContainer -
func (pcm *ProcessComponentsMock) FullArchiveInterceptorsContainer() process.InterceptorsContainer {
	return pcm.FullArchiveIntContainer
}

// ResolversContainer -
func (pcm *ProcessComponentsMock) ResolversContainer() dataRetriever.ResolversContainer {
	return pcm.ResContainer
}

// RequestersFinder -
func (pcm *ProcessComponentsMock) RequestersFinder() dataRetriever.RequestersFinder {
	return pcm.ReqFinder
}

// RoundHandler -
func (pcm *ProcessComponentsMock) RoundHandler() consensus.RoundHandler {
	return pcm.RoundHandlerField
}

// EpochStartTrigger -
func (pcm *ProcessComponentsMock) EpochStartTrigger() epochStart.TriggerHandler {
	return pcm.EpochTrigger
}

// EpochStartNotifier -
func (pcm *ProcessComponentsMock) EpochStartNotifier() factory.EpochStartNotifier {
	return pcm.EpochNotifier
}

// ForkDetector -
func (pcm *ProcessComponentsMock) ForkDetector() process.ForkDetector {
	return pcm.ForkDetect
}

// BlockProcessor -
func (pcm *ProcessComponentsMock) BlockProcessor() process.BlockProcessor {
	return pcm.BlockProcess
}

// BlackListHandler -
func (pcm *ProcessComponentsMock) BlackListHandler() process.TimeCacher {
	return pcm.BlackListHdl
}

// BootStorer -
func (pcm *ProcessComponentsMock) BootStorer() process.BootStorer {
	return pcm.BootSore
}

// HeaderSigVerifier -
func (pcm *ProcessComponentsMock) HeaderSigVerifier() process.InterceptedHeaderSigVerifier {
	return pcm.HeaderSigVerif
}

// HeaderIntegrityVerifier -
func (pcm *ProcessComponentsMock) HeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	return pcm.HeaderIntegrVerif
}

// ValidatorsStatistics -
func (pcm *ProcessComponentsMock) ValidatorsStatistics() process.ValidatorStatisticsProcessor {
	return pcm.ValidatorStatistics
}

// ValidatorsProvider -
func (pcm *ProcessComponentsMock) ValidatorsProvider() process.ValidatorsProvider {
	return pcm.ValidatorProvider
}

// BlockTracker -
func (pcm *ProcessComponentsMock) BlockTracker() process.BlockTracker {
	return pcm.BlockTrack
}

// PendingMiniBlocksHandler -
func (pcm *ProcessComponentsMock) PendingMiniBlocksHandler() process.PendingMiniBlocksHandler {
	return pcm.PendingMiniBlocksHdl
}

// RequestHandler -
func (pcm *ProcessComponentsMock) RequestHandler() process.RequestHandler {
	return pcm.ReqHandler
}

// TxLogsProcessor -
func (pcm *ProcessComponentsMock) TxLogsProcessor() process.TransactionLogProcessorDatabase {
	return pcm.TxLogsProcess
}

// HeaderConstructionValidator -
func (pcm *ProcessComponentsMock) HeaderConstructionValidator() process.HeaderConstructionValidator {
	return pcm.HeaderConstructValidator
}

// PeerShardMapper -
func (pcm *ProcessComponentsMock) PeerShardMapper() process.NetworkShardingCollector {
	return pcm.MainPeerMapper
}

// FullArchivePeerShardMapper -
func (pcm *ProcessComponentsMock) FullArchivePeerShardMapper() process.NetworkShardingCollector {
	return pcm.FullArchivePeerMapper
}

// FallbackHeaderValidator -
func (pcm *ProcessComponentsMock) FallbackHeaderValidator() process.FallbackHeaderValidator {
	return pcm.FallbackHdrValidator
}

// APITransactionEvaluator -
func (pcm *ProcessComponentsMock) APITransactionEvaluator() factory.TransactionEvaluator {
	return pcm.TransactionEvaluator
}

// WhiteListHandler -
func (pcm *ProcessComponentsMock) WhiteListHandler() process.WhiteListHandler {
	return pcm.WhiteListHandlerInternal
}

// WhiteListerVerifiedTxs -
func (pcm *ProcessComponentsMock) WhiteListerVerifiedTxs() process.WhiteListHandler {
	return pcm.WhiteListerVerifiedTxsInternal
}

// HistoryRepository -
func (pcm *ProcessComponentsMock) HistoryRepository() dblookupext.HistoryRepository {
	return pcm.HistoryRepositoryInternal
}

// ImportStartHandler -
func (pcm *ProcessComponentsMock) ImportStartHandler() update.ImportStartHandler {
	return pcm.ImportStartHandlerInternal
}

// RequestedItemsHandler -
func (pcm *ProcessComponentsMock) RequestedItemsHandler() dataRetriever.RequestedItemsHandler {
	return pcm.RequestedItemsHandlerInternal
}

// NodeRedundancyHandler -
func (pcm *ProcessComponentsMock) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	return pcm.NodeRedundancyHandlerInternal
}

// CurrentEpochProvider -
func (pcm *ProcessComponentsMock) CurrentEpochProvider() process.CurrentNetworkEpochProviderHandler {
	return pcm.CurrentEpochProviderInternal
}

// AccountsParser -
func (pcm *ProcessComponentsMock) AccountsParser() genesis.AccountsParser {
	return pcm.AccountsParserInternal
}

// String -
func (pcm *ProcessComponentsMock) String() string {
	return "ProcessComponentsMock"
}

// ScheduledTxsExecutionHandler -
func (pcm *ProcessComponentsMock) ScheduledTxsExecutionHandler() process.ScheduledTxsExecutionHandler {
	return pcm.ScheduledTxsExecutionHandlerInternal
}

// TxsSenderHandler -
func (pcm *ProcessComponentsMock) TxsSenderHandler() process.TxsSenderHandler {
	return pcm.TxsSenderHandlerField
}

// HardforkTrigger -
func (pcm *ProcessComponentsMock) HardforkTrigger() factory.HardforkTrigger {
	return pcm.HardforkTriggerField
}

// ProcessedMiniBlocksTracker -
func (pcm *ProcessComponentsMock) ProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker {
	return pcm.ProcessedMiniBlocksTrackerInternal
}

// ESDTDataStorageHandlerForAPI -
func (pcm *ProcessComponentsMock) ESDTDataStorageHandlerForAPI() vmcommon.ESDTNFTStorageHandler {
	return pcm.ESDTDataStorageHandlerForAPIInternal
}

// ReceiptsRepository -
func (pcm *ProcessComponentsMock) ReceiptsRepository() factory.ReceiptsRepository {
	return pcm.ReceiptsRepositoryInternal
}

// SentSignaturesTracker -
func (pcm *ProcessComponentsMock) SentSignaturesTracker() process.SentSignaturesTracker {
	return pcm.SentSignaturesTrackerInternal
}

// EpochSystemSCProcessor -
func (pcm *ProcessComponentsMock) EpochSystemSCProcessor() process.EpochStartSystemSCProcessor {
	return pcm.EpochSystemSCProcessorInternal
}

// IsInterfaceNil -
func (pcm *ProcessComponentsMock) IsInterfaceNil() bool {
	return pcm == nil
}
