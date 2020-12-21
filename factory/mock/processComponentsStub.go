package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ProcessComponentsMock -
type ProcessComponentsMock struct {
	NodesCoord                     sharding.NodesCoordinator
	ShardCoord                     sharding.Coordinator
	IntContainer                   process.InterceptorsContainer
	ResFinder                      dataRetriever.ResolversFinder
	RoundHandlerField              consensus.RoundHandler
	EpochTrigger                   epochStart.TriggerHandler
	EpochNotifier                  factory.EpochStartNotifier
	ForkDetect                     process.ForkDetector
	BlockProcess                   process.BlockProcessor
	BlackListHdl                   process.TimeCacher
	BootSore                       process.BootStorer
	HeaderSigVerif                 process.InterceptedHeaderSigVerifier
	HeaderIntegrVerif              process.HeaderIntegrityVerifier
	ValidatorStatistics            process.ValidatorStatisticsProcessor
	ValidatorProvider              process.ValidatorsProvider
	BlockTrack                     process.BlockTracker
	PendingMiniBlocksHdl           process.PendingMiniBlocksHandler
	ReqHandler                     process.RequestHandler
	TxLogsProcess                  process.TransactionLogProcessorDatabase
	HeaderConstructValidator       process.HeaderConstructionValidator
	PeerMapper                     process.NetworkShardingCollector
	TxSimulatorProcessor           factory.TransactionSimulatorProcessor
	FallbackHdrValidator           process.FallbackHeaderValidator
	WhiteListHandlerInternal       process.WhiteListHandler
	WhiteListerVerifiedTxsInternal process.WhiteListHandler
	HistoryRepositoryInternal      dblookupext.HistoryRepository
	ImportStartHandlerInternal     update.ImportStartHandler
	RequestedItemsHandlerInternal  dataRetriever.RequestedItemsHandler
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
func (pcm *ProcessComponentsMock) NodesCoordinator() sharding.NodesCoordinator {
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

// ResolversFinder -
func (pcm *ProcessComponentsMock) ResolversFinder() dataRetriever.ResolversFinder {
	return pcm.ResFinder
}

// RoundHandlerField -
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
	return pcm.PeerMapper
}

// FallbackHeaderValidator -
func (pcm *ProcessComponentsMock) FallbackHeaderValidator() process.FallbackHeaderValidator {
	return pcm.FallbackHdrValidator
}

// TransactionSimulatorProcessor -
func (pcm *ProcessComponentsMock) TransactionSimulatorProcessor() factory.TransactionSimulatorProcessor {
	return pcm.TxSimulatorProcessor
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

// IsInterfaceNil -
func (pcm *ProcessComponentsMock) IsInterfaceNil() bool {
	return pcm == nil
}
