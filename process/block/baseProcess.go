package block

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/logging"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	debugFactory "github.com/multiversx/mx-chain-go/debug/factory"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

var log = logger.GetOrCreate("process/block")

// CrossShardIncomingMbsCreationResult represents the result of creating cross-shard mini blocks
type CrossShardIncomingMbsCreationResult struct {
	HeaderFinished    bool
	PendingMiniBlocks []block.MiniblockAndHash
	AddedMiniBlocks   []block.MiniblockAndHash
}

type hashAndHdr struct {
	hdr  data.HeaderHandler
	hash []byte
}

type splitTxsResult struct {
	incomingMiniBlocks        []data.MiniBlockHeaderHandler
	outGoingMiniBlocks        []data.MiniBlockHeaderHandler
	incomingTransactions      map[string][]data.TransactionHandler
	outgoingTransactionHashes [][]byte
	outgoingTransactions      []data.TransactionHandler
}

type baseProcessor struct {
	shardCoordinator        sharding.Coordinator
	nodesCoordinator        nodesCoordinator.NodesCoordinator
	accountsDB              map[state.AccountsDbIdentifier]state.AccountsAdapter
	accountsProposal        state.AccountsAdapter
	forkDetector            process.ForkDetector
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	store                   dataRetriever.StorageService
	uint64Converter         typeConverters.Uint64ByteSliceConverter
	blockSizeThrottler      process.BlockSizeThrottler
	epochStartTrigger       process.EpochStartTriggerHandler
	headerValidator         process.HeaderConstructionValidator
	blockChainHook          process.BlockChainHookHandler
	txCoordinator           process.TransactionCoordinator
	roundHandler            consensus.RoundHandler
	bootStorer              process.BootStorer
	requestBlockBodyHandler process.RequestBlockBodyHandler
	requestHandler          process.RequestHandler
	blockTracker            process.BlockTracker
	dataPool                dataRetriever.PoolsHolder
	feeHandler              process.TransactionFeeHandler
	blockChain              data.ChainHandler
	hdrsForCurrBlock        HeadersForBlock
	genesisNonce            uint64
	mutProcessDebugger      sync.RWMutex
	processDebugger         process.Debugger
	processStatusHandler    common.ProcessStatusHandler
	managedPeersHolder      common.ManagedPeersHolder
	sentSignaturesTracker   process.SentSignaturesTracker

	versionedHeaderFactory       nodeFactory.VersionedHeaderFactory
	headerIntegrityVerifier      process.HeaderIntegrityVerifier
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	blockProcessingCutoffHandler cutoff.BlockProcessingCutoffHandler

	appStatusHandler core.AppStatusHandler
	blockProcessor   blockProcessor
	txCounter        *transactionCounter

	outportHandler                outport.OutportHandler
	outportDataProvider           outport.DataProviderOutport
	historyRepo                   dblookupext.HistoryRepository
	epochNotifier                 process.EpochNotifier
	enableEpochsHandler           common.EnableEpochsHandler
	roundNotifier                 process.RoundNotifier
	enableRoundsHandler           common.EnableRoundsHandler
	vmContainerFactory            process.VirtualMachinesContainerFactory
	vmContainer                   process.VirtualMachinesContainer
	gasConsumedProvider           gasConsumedProvider
	economicsData                 process.EconomicsDataHandler
	epochChangeGracePeriodHandler common.EpochChangeGracePeriodHandler
	stateAccessesCollector        state.StateAccessesCollector
	processConfigsHandler         common.ProcessConfigsHandler

	processDataTriesOnCommitEpoch bool
	lastRestartNonce              uint64
	pruningDelay                  uint32
	processedMiniBlocksTracker    process.ProcessedMiniBlocksTracker
	receiptsRepository            receiptsRepository

	mutNonceOfFirstCommittedBlock sync.RWMutex
	nonceOfFirstCommittedBlock    core.OptionalUint64

	proofsPool                         dataRetriever.ProofsPool
	executionResultsInclusionEstimator process.InclusionEstimator
	miniBlocksSelectionSession         MiniBlocksSelectionSession
	executionResultsVerifier           ExecutionResultsVerifier
	missingDataResolver                MissingDataResolver
	gasComputation                     process.GasComputation
	executionManager                   process.ExecutionManager
}

type bootStorerDataArgs struct {
	headerInfo                 bootstrapStorage.BootstrapHeaderInfo
	lastSelfNotarizedHeaders   []bootstrapStorage.BootstrapHeaderInfo
	round                      uint64
	highestFinalBlockNonce     uint64
	pendingMiniBlocks          []bootstrapStorage.PendingMiniBlocksInfo
	processedMiniBlocks        []bootstrapStorage.MiniBlocksInMeta
	nodesCoordinatorConfigKey  []byte
	epochStartTriggerConfigKey []byte
}

// NewBaseProcessor will create a new instance of baseProcessor
func NewBaseProcessor(arguments ArgBaseProcessor) (*baseProcessor, error) {
	err := checkProcessorParameters(arguments)
	if err != nil {
		return nil, err
	}

	processDebugger, err := createDisabledProcessDebugger()
	if err != nil {
		return nil, err
	}

	genesisHdr := arguments.DataComponents.Blockchain().GetGenesisHeader()
	if check.IfNil(genesisHdr) {
		return nil, fmt.Errorf("%w for genesis header in DataComponents.Blockchain", process.ErrNilHeaderHandler)
	}

	base := &baseProcessor{
		accountsDB:                         arguments.AccountsDB,
		accountsProposal:                   arguments.AccountsProposal,
		blockSizeThrottler:                 arguments.BlockSizeThrottler,
		forkDetector:                       arguments.ForkDetector,
		hasher:                             arguments.CoreComponents.Hasher(),
		marshalizer:                        arguments.CoreComponents.InternalMarshalizer(),
		store:                              arguments.DataComponents.StorageService(),
		shardCoordinator:                   arguments.BootstrapComponents.ShardCoordinator(),
		feeHandler:                         arguments.FeeHandler,
		nodesCoordinator:                   arguments.NodesCoordinator,
		uint64Converter:                    arguments.CoreComponents.Uint64ByteSliceConverter(),
		requestHandler:                     arguments.RequestHandler,
		appStatusHandler:                   arguments.StatusCoreComponents.AppStatusHandler(),
		blockChainHook:                     arguments.BlockChainHook,
		txCoordinator:                      arguments.TxCoordinator,
		epochStartTrigger:                  arguments.EpochStartTrigger,
		headerValidator:                    arguments.HeaderValidator,
		roundHandler:                       arguments.CoreComponents.RoundHandler(),
		bootStorer:                         arguments.BootStorer,
		blockTracker:                       arguments.BlockTracker,
		dataPool:                           arguments.DataComponents.Datapool(),
		blockChain:                         arguments.DataComponents.Blockchain(),
		outportHandler:                     arguments.StatusComponents.OutportHandler(),
		genesisNonce:                       genesisHdr.GetNonce(),
		versionedHeaderFactory:             arguments.BootstrapComponents.VersionedHeaderFactory(),
		headerIntegrityVerifier:            arguments.BootstrapComponents.HeaderIntegrityVerifier(),
		historyRepo:                        arguments.HistoryRepository,
		epochNotifier:                      arguments.CoreComponents.EpochNotifier(),
		enableEpochsHandler:                arguments.CoreComponents.EnableEpochsHandler(),
		roundNotifier:                      arguments.CoreComponents.RoundNotifier(),
		enableRoundsHandler:                arguments.CoreComponents.EnableRoundsHandler(),
		epochChangeGracePeriodHandler:      arguments.CoreComponents.EpochChangeGracePeriodHandler(),
		vmContainerFactory:                 arguments.VMContainersFactory,
		vmContainer:                        arguments.VmContainer,
		processDataTriesOnCommitEpoch:      arguments.Config.Debug.EpochStart.ProcessDataTrieOnCommitEpoch,
		gasConsumedProvider:                arguments.GasHandler,
		economicsData:                      arguments.CoreComponents.EconomicsData(),
		scheduledTxsExecutionHandler:       arguments.ScheduledTxsExecutionHandler,
		pruningDelay:                       pruningDelay,
		processedMiniBlocksTracker:         arguments.ProcessedMiniBlocksTracker,
		receiptsRepository:                 arguments.ReceiptsRepository,
		processDebugger:                    processDebugger,
		outportDataProvider:                arguments.OutportDataProvider,
		processStatusHandler:               arguments.CoreComponents.ProcessStatusHandler(),
		blockProcessingCutoffHandler:       arguments.BlockProcessingCutoffHandler,
		managedPeersHolder:                 arguments.ManagedPeersHolder,
		sentSignaturesTracker:              arguments.SentSignaturesTracker,
		stateAccessesCollector:             arguments.StateAccessesCollector,
		proofsPool:                         arguments.DataComponents.Datapool().Proofs(),
		hdrsForCurrBlock:                   arguments.HeadersForBlock,
		processConfigsHandler:              arguments.CoreComponents.ProcessConfigsHandler(),
		executionResultsInclusionEstimator: arguments.ExecutionResultsInclusionEstimator,
		miniBlocksSelectionSession:         arguments.MiniBlocksSelectionSession,
		executionResultsVerifier:           arguments.ExecutionResultsVerifier,
		missingDataResolver:                arguments.MissingDataResolver,
		gasComputation:                     arguments.GasComputation,
		executionManager:                   arguments.ExecutionManager,
	}

	return base, nil
}

func checkForNils(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}
	if check.IfNil(bodyHandler) {
		return process.ErrNilBlockBody
	}
	return nil
}

// checkBlockValidity method checks if the given block is valid
func (bp *baseProcessor) checkBlockValidity(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	err := checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	currentBlockHeader := bp.blockChain.GetCurrentBlockHeader()

	if check.IfNil(currentBlockHeader) {
		if headerHandler.GetNonce() == bp.genesisNonce+1 { // first block after genesis
			if bytes.Equal(headerHandler.GetPrevHash(), bp.blockChain.GetGenesisHeaderHash()) {
				// TODO: add genesis block verification
				return nil
			}

			log.Debug("hash does not match",
				"local block hash", bp.blockChain.GetGenesisHeaderHash(),
				"received previous hash", headerHandler.GetPrevHash())

			return process.ErrBlockHashDoesNotMatch
		}

		log.Debug("nonce does not match",
			"local block nonce", 0,
			"received nonce", headerHandler.GetNonce())

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetRound() <= currentBlockHeader.GetRound() {
		log.Debug("round does not match",
			"local block round", currentBlockHeader.GetRound(),
			"received block round", headerHandler.GetRound())

		return process.ErrLowerRoundInBlock
	}

	if headerHandler.GetNonce() != currentBlockHeader.GetNonce()+1 {
		log.Debug("nonce does not match",
			"local block nonce", currentBlockHeader.GetNonce(),
			"received nonce", headerHandler.GetNonce())

		return process.ErrWrongNonceInBlock
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), bp.blockChain.GetCurrentBlockHeaderHash()) {
		log.Debug("hash does not match",
			"local block hash", bp.blockChain.GetCurrentBlockHeaderHash(),
			"received previous hash", headerHandler.GetPrevHash())

		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(headerHandler.GetPrevRandSeed(), currentBlockHeader.GetRandSeed()) {
		log.Debug("random seed does not match",
			"local random seed", currentBlockHeader.GetRandSeed(),
			"received previous random seed", headerHandler.GetPrevRandSeed())

		return process.ErrRandSeedDoesNotMatch
	}

	// verification of epoch
	if headerHandler.GetEpoch() < currentBlockHeader.GetEpoch() {
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

// checkScheduledRootHash checks if the scheduled root hash from the given header is the same with the current user accounts state root hash
func (bp *baseProcessor) checkScheduledRootHash(headerHandler data.HeaderHandler) error {
	if !bp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return nil
	}

	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}

	additionalData := headerHandler.GetAdditionalData()
	if check.IfNil(additionalData) {
		return process.ErrNilAdditionalData
	}

	if !bytes.Equal(additionalData.GetScheduledRootHash(), bp.getRootHash()) {
		log.Debug("scheduled root hash does not match",
			"current root hash", bp.getRootHash(),
			"header scheduled root hash", additionalData.GetScheduledRootHash())
		return process.ErrScheduledRootHashDoesNotMatch
	}

	return nil
}

// verifyStateRoot verifies the state root hash given as parameter against the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *baseProcessor) verifyStateRoot(rootHash []byte) bool {
	trieRootHash, err := bp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		log.Debug("verify account.RootHash", "error", err.Error())
	}

	return bytes.Equal(trieRootHash, rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	rootHash, err := bp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		log.Trace("get account.RootHash", "error", err.Error())
	}

	return rootHash
}

func (bp *baseProcessor) requestHeadersIfMissing(
	sortedHdrs []data.HeaderHandler,
	shardId uint32,
) error {

	prevHdr, _, err := bp.blockTracker.GetLastCrossNotarizedHeader(shardId)
	if err != nil {
		return err
	}

	lastNotarizedHdrEpoch := prevHdr.GetEpoch()
	lastNotarizedHdrRound := prevHdr.GetRound()
	lastNotarizedHdrNonce := prevHdr.GetNonce()

	missingNonces := make([]uint64, 0)
	for i := 0; i < len(sortedHdrs); i++ {
		currHdr := sortedHdrs[i]
		if check.IfNil(currHdr) {
			continue
		}

		hdrTooOld := currHdr.GetRound() <= lastNotarizedHdrRound
		if hdrTooOld {
			continue
		}

		maxNumNoncesToAdd := process.MaxHeaderRequestsAllowed - int(int64(prevHdr.GetNonce())-int64(lastNotarizedHdrNonce))
		if maxNumNoncesToAdd <= 0 {
			break
		}

		noncesDiff := int64(currHdr.GetNonce()) - int64(prevHdr.GetNonce())
		nonces := addMissingNonces(noncesDiff, prevHdr.GetNonce(), maxNumNoncesToAdd)
		missingNonces = append(missingNonces, nonces...)

		prevHdr = currHdr
	}

	maxNumNoncesToAdd := process.MaxHeaderRequestsAllowed - int(int64(prevHdr.GetNonce())-int64(lastNotarizedHdrNonce))
	if maxNumNoncesToAdd > 0 {
		lastRound := bp.roundHandler.Index() - 1
		roundsDiff := lastRound - int64(prevHdr.GetRound())
		nonces := addMissingNonces(roundsDiff, prevHdr.GetNonce(), maxNumNoncesToAdd)
		missingNonces = append(missingNonces, nonces...)
	}

	for _, nonce := range missingNonces {
		bp.addHeaderIntoTrackerPool(nonce, shardId)
		go bp.requestHeaderIfNeeded(nonce, shardId)
		go bp.requestProofIfNeeded(nonce, shardId, lastNotarizedHdrEpoch)
	}

	return nil
}

func addMissingNonces(diff int64, lastNonce uint64, maxNumNoncesToAdd int) []uint64 {
	missingNonces := make([]uint64, 0)

	if diff < 2 {
		return missingNonces
	}

	numNonces := uint64(diff) - 1
	startNonce := lastNonce + 1
	endNonce := startNonce + numNonces

	for nonce := startNonce; nonce < endNonce; nonce++ {
		missingNonces = append(missingNonces, nonce)
		if len(missingNonces) >= maxNumNoncesToAdd {
			break
		}
	}

	return missingNonces
}

func displayHeader(
	headerHandler data.HeaderHandler,
	headerProof data.HeaderProofHandler,
) []*display.LineData {
	var valStatRootHash, epochStartMetaHash, scheduledRootHash []byte
	metaHeader, isMetaHeader := headerHandler.(data.MetaHeaderHandler)
	if isMetaHeader {
		valStatRootHash = metaHeader.GetValidatorStatsRootHash()
	} else {
		shardHeader, isShardHeader := headerHandler.(data.ShardHeaderHandler)
		if isShardHeader {
			epochStartMetaHash = shardHeader.GetEpochStartMetaHash()
		}
	}
	additionalData := headerHandler.GetAdditionalData()
	if !check.IfNil(additionalData) {
		scheduledRootHash = additionalData.GetScheduledRootHash()
	}

	var aggregatedSig, bitmap []byte
	var proofShard, proofEpoch uint32
	var proofRound, proofNonce uint64
	var isStartOfEpoch, hasProofInfo bool
	if !check.IfNil(headerProof) {
		hasProofInfo = true
		aggregatedSig, bitmap = headerProof.GetAggregatedSignature(), headerProof.GetPubKeysBitmap()
		proofShard = headerProof.GetHeaderShardId()
		proofEpoch = headerProof.GetHeaderEpoch()
		proofRound = headerProof.GetHeaderRound()
		proofNonce = headerProof.GetHeaderNonce()
		isStartOfEpoch = headerProof.GetIsStartOfEpoch()
	}

	logLines := []*display.LineData{
		display.NewLineData(false, []string{
			"",
			"ChainID",
			logger.DisplayByteSlice(headerHandler.GetChainID())}),
		display.NewLineData(false, []string{
			"",
			"Epoch",
			fmt.Sprintf("%d", headerHandler.GetEpoch())}),
		display.NewLineData(false, []string{
			"",
			"Round",
			fmt.Sprintf("%d", headerHandler.GetRound())}),
		display.NewLineData(false, []string{
			"",
			"TimeStamp",
			fmt.Sprintf("%d", headerHandler.GetTimeStamp())}),
		display.NewLineData(false, []string{
			"",
			"Nonce",
			fmt.Sprintf("%d", headerHandler.GetNonce())}),
		display.NewLineData(false, []string{
			"",
			"Prev hash",
			logger.DisplayByteSlice(headerHandler.GetPrevHash())}),
		display.NewLineData(false, []string{
			"",
			"Prev rand seed",
			logger.DisplayByteSlice(headerHandler.GetPrevRandSeed())}),
		display.NewLineData(false, []string{
			"",
			"Rand seed",
			logger.DisplayByteSlice(headerHandler.GetRandSeed())}),
		display.NewLineData(false, []string{
			"",
			"Pub keys bitmap",
			hex.EncodeToString(headerHandler.GetPubKeysBitmap())}),
		display.NewLineData(false, []string{
			"",
			"Signature",
			logger.DisplayByteSlice(headerHandler.GetSignature())}),
		display.NewLineData(false, []string{
			"",
			"Leader's Signature",
			logger.DisplayByteSlice(headerHandler.GetLeaderSignature())}),
		display.NewLineData(false, []string{
			"",
			"Scheduled root hash",
			logger.DisplayByteSlice(scheduledRootHash)}),
		display.NewLineData(false, []string{
			"",
			"Root hash",
			logger.DisplayByteSlice(headerHandler.GetRootHash())}),
		display.NewLineData(false, []string{
			"",
			"Validator stats root hash",
			logger.DisplayByteSlice(valStatRootHash)}),
		display.NewLineData(false, []string{
			"",
			"Receipts hash",
			logger.DisplayByteSlice(headerHandler.GetReceiptsHash())}),
		display.NewLineData(true, []string{
			"",
			"Epoch start meta hash",
			logger.DisplayByteSlice(epochStartMetaHash)}),
	}

	if hasProofInfo {
		logLines = append(logLines,
			display.NewLineData(false, []string{
				"Header proof",
				"Aggregated signature",
				logger.DisplayByteSlice(aggregatedSig)}),
			display.NewLineData(false, []string{
				"",
				"Pub keys bitmap",
				logger.DisplayByteSlice(bitmap)}),
			display.NewLineData(false, []string{
				"",
				"Epoch",
				fmt.Sprintf("%d", proofEpoch)}),
			display.NewLineData(false, []string{
				"",
				"Round",
				fmt.Sprintf("%d", proofRound)}),
			display.NewLineData(false, []string{
				"",
				"Shard",
				fmt.Sprintf("%d", proofShard)}),
			display.NewLineData(false, []string{
				"",
				"Nonce",
				fmt.Sprintf("%d", proofNonce)}),
			display.NewLineData(true, []string{
				"",
				"IsStartOfEpoch",
				fmt.Sprintf("%t", isStartOfEpoch)}),
		)
	}

	return logLines
}

// checkProcessorParameters will check the input parameters values
func checkProcessorParameters(arguments ArgBaseProcessor) error {
	for key := range arguments.AccountsDB {
		if check.IfNil(arguments.AccountsDB[key]) {
			return process.ErrNilAccountsAdapter
		}
	}
	if check.IfNil(arguments.AccountsProposal) {
		return fmt.Errorf("%w for proposal", process.ErrNilAccountsAdapter)
	}
	if check.IfNil(arguments.DataComponents) {
		return process.ErrNilDataComponentsHolder
	}
	if check.IfNil(arguments.CoreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(arguments.BootstrapComponents) {
		return process.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(arguments.StatusComponents) {
		return process.ErrNilStatusComponentsHolder
	}
	if check.IfNil(arguments.ForkDetector) {
		return process.ErrNilForkDetector
	}
	if check.IfNil(arguments.CoreComponents.Hasher()) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.CoreComponents.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.DataComponents.StorageService()) {
		return process.ErrNilStorage
	}
	if check.IfNil(arguments.BootstrapComponents.ShardCoordinator()) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arguments.CoreComponents.Uint64ByteSliceConverter()) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(arguments.EpochStartTrigger) {
		return process.ErrNilEpochStartTrigger
	}
	if check.IfNil(arguments.CoreComponents.RoundHandler()) {
		return process.ErrNilRoundHandler
	}
	if check.IfNil(arguments.BootStorer) {
		return process.ErrNilStorage
	}
	if check.IfNil(arguments.BlockChainHook) {
		return process.ErrNilBlockChainHook
	}
	if check.IfNil(arguments.TxCoordinator) {
		return process.ErrNilTransactionCoordinator
	}
	if check.IfNil(arguments.HeaderValidator) {
		return process.ErrNilHeaderValidator
	}
	if check.IfNil(arguments.BlockTracker) {
		return process.ErrNilBlockTracker
	}
	if check.IfNil(arguments.FeeHandler) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(arguments.DataComponents.Blockchain()) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(arguments.BlockSizeThrottler) {
		return process.ErrNilBlockSizeThrottler
	}
	if check.IfNil(arguments.StatusComponents.OutportHandler()) {
		return process.ErrNilOutportHandler
	}
	if check.IfNil(arguments.HistoryRepository) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(arguments.BootstrapComponents.HeaderIntegrityVerifier()) {
		return process.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(arguments.CoreComponents.EpochNotifier()) {
		return process.ErrNilEpochNotifier
	}
	enableEpochsHandler := arguments.CoreComponents.EnableEpochsHandler()
	if check.IfNil(enableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(enableEpochsHandler, []core.EnableEpochFlag{
		common.ScheduledMiniBlocksFlag,
		common.StakingV2Flag,
		common.CurrentRandomnessOnSortingFlag,
		common.AndromedaFlag,
	})
	if err != nil {
		return err
	}
	if check.IfNil(arguments.CoreComponents.EpochChangeGracePeriodHandler()) {
		return process.ErrNilEpochChangeGracePeriodHandler
	}
	if check.IfNil(arguments.CoreComponents.RoundNotifier()) {
		return process.ErrNilRoundNotifier
	}
	if check.IfNil(arguments.CoreComponents.EnableRoundsHandler()) {
		return process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(arguments.StatusCoreComponents) {
		return process.ErrNilStatusCoreComponentsHolder
	}
	if check.IfNil(arguments.StatusCoreComponents.AppStatusHandler()) {
		return process.ErrNilAppStatusHandler
	}
	if check.IfNil(arguments.GasHandler) {
		return process.ErrNilGasHandler
	}
	if check.IfNil(arguments.CoreComponents.EconomicsData()) {
		return process.ErrNilEconomicsData
	}
	if check.IfNil(arguments.OutportDataProvider) {
		return process.ErrNilOutportDataProvider
	}
	if check.IfNil(arguments.ScheduledTxsExecutionHandler) {
		return process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(arguments.BootstrapComponents.VersionedHeaderFactory()) {
		return process.ErrNilVersionedHeaderFactory
	}
	if check.IfNil(arguments.ProcessedMiniBlocksTracker) {
		return process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(arguments.ReceiptsRepository) {
		return process.ErrNilReceiptsRepository
	}
	if check.IfNil(arguments.BlockProcessingCutoffHandler) {
		return process.ErrNilBlockProcessingCutoffHandler
	}
	if check.IfNil(arguments.ManagedPeersHolder) {
		return process.ErrNilManagedPeersHolder
	}
	if check.IfNil(arguments.SentSignaturesTracker) {
		return process.ErrNilSentSignatureTracker
	}
	if check.IfNil(arguments.StateAccessesCollector) {
		return process.ErrNilStateAccessesCollector
	}
	if check.IfNil(arguments.HeadersForBlock) {
		return process.ErrNilHeadersForBlock
	}
	if check.IfNil(arguments.DataComponents.Datapool()) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(arguments.DataComponents.Datapool().Headers()) {
		return process.ErrNilHeadersDataPool
	}
	if check.IfNil(arguments.ExecutionResultsInclusionEstimator) {
		return process.ErrNilExecutionResultsInclusionEstimator
	}
	if check.IfNil(arguments.MiniBlocksSelectionSession) {
		return process.ErrNilMiniBlocksSelectionSession
	}
	if check.IfNil(arguments.ExecutionResultsVerifier) {
		return process.ErrNilExecutionResultsVerifier
	}
	if check.IfNil(arguments.MissingDataResolver) {
		return process.ErrNilMissingDataResolver
	}
	if check.IfNil(arguments.GasComputation) {
		return process.ErrNilGasComputation
	}
	if check.IfNil(arguments.ExecutionManager) {
		return process.ErrNilExecutionManager
	}

	return nil
}

func (bp *baseProcessor) createBlockStarted() error {
	bp.hdrsForCurrBlock.Reset()
	scheduledGasAndFees := bp.scheduledTxsExecutionHandler.GetScheduledGasAndFees()
	bp.txCoordinator.CreateBlockStarted()
	bp.feeHandler.CreateBlockStarted(scheduledGasAndFees)

	err := bp.txCoordinator.AddIntermediateTransactions(bp.scheduledTxsExecutionHandler.GetScheduledIntermediateTxs(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (bp *baseProcessor) verifyFees(header data.HeaderHandler) error {
	if header.GetAccumulatedFees().Cmp(bp.feeHandler.GetAccumulatedFees()) != 0 {
		return process.ErrAccumulatedFeesDoNotMatch
	}
	if header.GetDeveloperFees().Cmp(bp.feeHandler.GetDeveloperFees()) != 0 {
		return process.ErrDeveloperFeesDoNotMatch
	}

	return nil
}

// TODO: remove bool parameter and give instead the set to sort
func (bp *baseProcessor) sortHeadersForCurrentBlockByNonce(usedInBlock bool) (map[uint32][]data.HeaderHandler, error) {
	hdrsForCurrentBlock, err := bp.hdrsForCurrBlock.ComputeHeadersForCurrentBlock(usedInBlock)
	if err != nil {
		return nil, err
	}

	// sort headers for each shard
	for _, hdrsForShard := range hdrsForCurrentBlock {
		process.SortHeadersByNonce(hdrsForShard)
	}

	return hdrsForCurrentBlock, nil
}

func (bp *baseProcessor) sortHeaderHashesForCurrentBlockByNonce(usedInBlock bool) (map[uint32][][]byte, error) {
	hdrsForCurrentBlockInfo, err := bp.hdrsForCurrBlock.ComputeHeadersForCurrentBlockInfo(usedInBlock)
	if err != nil {
		return nil, err
	}

	for _, hdrsForShard := range hdrsForCurrentBlockInfo {
		if len(hdrsForShard) > 1 {
			sort.Slice(hdrsForShard, func(i, j int) bool {
				return hdrsForShard[i].GetNonce() < hdrsForShard[j].GetNonce()
			})
		}
	}

	hdrsHashesForCurrentBlock := make(map[uint32][][]byte, len(hdrsForCurrentBlockInfo))
	for shardId, hdrsForShard := range hdrsForCurrentBlockInfo {
		for _, hdrForShard := range hdrsForShard {
			hdrsHashesForCurrentBlock[shardId] = append(hdrsHashesForCurrentBlock[shardId], hdrForShard.GetHash())
		}
	}

	return hdrsHashesForCurrentBlock, nil
}

func (bp *baseProcessor) createMiniBlockHeaderHandlers(
	body *block.Body,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (int, []data.MiniBlockHeaderHandler, error) {
	if len(body.MiniBlocks) == 0 {
		return 0, nil, nil
	}

	totalTxCount := 0
	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(body.MiniBlocks))

	for i := 0; i < len(body.MiniBlocks); i++ {
		txCount := len(body.MiniBlocks[i].TxHashes)
		totalTxCount += txCount

		miniBlockHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, body.MiniBlocks[i])
		if err != nil {
			return 0, nil, err
		}

		miniBlockHeaderHandlers[i] = &block.MiniBlockHeader{
			Hash:            miniBlockHash,
			SenderShardID:   body.MiniBlocks[i].SenderShardID,
			ReceiverShardID: body.MiniBlocks[i].ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            body.MiniBlocks[i].Type,
		}

		err = bp.setMiniBlockHeaderReservedField(body.MiniBlocks[i], miniBlockHeaderHandlers[i], processedMiniBlocksDestMeInfo)
		if err != nil {
			return 0, nil, err
		}
	}

	return totalTxCount, miniBlockHeaderHandlers, nil
}

func (bp *baseProcessor) setMiniBlockHeaderReservedField(
	miniBlock *block.MiniBlock,
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	if !bp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return nil
	}

	err := bp.setIndexOfFirstTxProcessed(miniBlockHeaderHandler)
	if err != nil {
		return err
	}

	err = bp.setIndexOfLastTxProcessed(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
	if err != nil {
		return err
	}

	notEmpty := len(miniBlock.TxHashes) > 0
	isScheduledMiniBlock := notEmpty && bp.scheduledTxsExecutionHandler.IsScheduledTx(miniBlock.TxHashes[0])
	if isScheduledMiniBlock {
		return bp.setProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
	}

	return bp.setProcessingTypeAndConstructionStateForNormalMb(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
}

func (bp *baseProcessor) setIndexOfFirstTxProcessed(miniBlockHeaderHandler data.MiniBlockHeaderHandler) error {
	processedMiniBlockInfo, _ := bp.processedMiniBlocksTracker.GetProcessedMiniBlockInfo(miniBlockHeaderHandler.GetHash())
	return miniBlockHeaderHandler.SetIndexOfFirstTxProcessed(processedMiniBlockInfo.IndexOfLastTxProcessed + 1)
}

func (bp *baseProcessor) setIndexOfLastTxProcessed(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	processedMiniBlockInfo := processedMiniBlocksDestMeInfo[string(miniBlockHeaderHandler.GetHash())]
	if processedMiniBlockInfo != nil {
		return miniBlockHeaderHandler.SetIndexOfLastTxProcessed(processedMiniBlockInfo.IndexOfLastTxProcessed)
	}

	return miniBlockHeaderHandler.SetIndexOfLastTxProcessed(int32(miniBlockHeaderHandler.GetTxCount()) - 1)
}

func (bp *baseProcessor) setProcessingTypeAndConstructionStateForScheduledMb(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	err := miniBlockHeaderHandler.SetProcessingType(int32(block.Scheduled))
	if err != nil {
		return err
	}

	if miniBlockHeaderHandler.GetSenderShardID() == bp.shardCoordinator.SelfId() {
		err = miniBlockHeaderHandler.SetConstructionState(int32(block.Proposed))
		if err != nil {
			return err
		}
	} else {
		constructionState := getConstructionState(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
		err = miniBlockHeaderHandler.SetConstructionState(constructionState)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bp *baseProcessor) setProcessingTypeAndConstructionStateForNormalMb(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	if bp.scheduledTxsExecutionHandler.IsMiniBlockExecuted(miniBlockHeaderHandler.GetHash()) {
		err := miniBlockHeaderHandler.SetProcessingType(int32(block.Processed))
		if err != nil {
			return err
		}
	} else {
		err := miniBlockHeaderHandler.SetProcessingType(int32(block.Normal))
		if err != nil {
			return err
		}
	}

	constructionState := getConstructionState(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
	err := miniBlockHeaderHandler.SetConstructionState(constructionState)
	if err != nil {
		return err
	}

	return nil
}

func getConstructionState(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) int32 {
	constructionState := int32(block.Final)
	if isPartiallyExecuted(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo) {
		constructionState = int32(block.PartialExecuted)
	}

	return constructionState
}

func isPartiallyExecuted(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) bool {
	processedMiniBlockInfo := processedMiniBlocksDestMeInfo[string(miniBlockHeaderHandler.GetHash())]
	return processedMiniBlockInfo != nil && !processedMiniBlockInfo.FullyProcessed
}

// check if header has the same mini blocks as presented in body
func (bp *baseProcessor) checkHeaderBodyCorrelationProposal(miniBlockHeaders []data.MiniBlockHeaderHandler, body *block.Body) error {
	if len(miniBlockHeaders) != len(body.MiniBlocks) {
		return process.ErrHeaderBodyMismatch
	}

	var mbHdr data.MiniBlockHeaderHandler
	var miniBlock *block.MiniBlock
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock = body.MiniBlocks[i]
		mbHdr = miniBlockHeaders[i]
		if miniBlock == nil {
			return process.ErrNilMiniBlock
		}
		if mbHdr == nil {
			return process.ErrNilMiniBlockHeader
		}

		mbHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, miniBlock)
		if err != nil {
			return err
		}

		err = checkMiniBlockWithMiniBlockHeader(mbHash, mbHdr, miniBlock)
		if err != nil {
			return err
		}
	}

	return bp.checkMiniBlocksConstructionProposal(miniBlockHeaders)
}

func (bp *baseProcessor) checkMiniBlocksConstructionProposal(miniBlockHeaders []data.MiniBlockHeaderHandler) error {
	for i := 0; i < len(miniBlockHeaders); i++ {
		// for Supernova all miniBlocks not part of an execution result need to have construction state Proposed
		if miniBlockHeaders[i].GetConstructionState() != int32(block.Proposed) {
			return process.ErrWrongMiniBlockConstructionState
		}
		if miniBlockHeaders[i].GetProcessingType() != int32(block.Normal) {
			return process.ErrWrongMiniBlockProcessingType
		}
	}
	return nil
}

func checkMiniBlockWithMiniBlockHeader(mbHash []byte, mbHdr data.MiniBlockHeaderHandler, miniBlock *block.MiniBlock) error {
	if !bytes.Equal(mbHash, mbHdr.GetHash()) {
		return process.ErrHeaderBodyMismatch
	}

	if mbHdr.GetTxCount() != uint32(len(miniBlock.TxHashes)) {
		return process.ErrHeaderBodyMismatch
	}

	if mbHdr.GetReceiverShardID() != miniBlock.ReceiverShardID {
		return fmt.Errorf("%w: different mb receiver shard ID", process.ErrHeaderBodyMismatch)
	}

	if mbHdr.GetSenderShardID() != miniBlock.SenderShardID {
		return fmt.Errorf("%w: different mb sender shard ID", process.ErrHeaderBodyMismatch)
	}
	return nil
}

// check if header has the same mini blocks as presented in body
func (bp *baseProcessor) checkHeaderBodyCorrelation(miniBlockHeaders []data.MiniBlockHeaderHandler, body *block.Body) error {
	if len(miniBlockHeaders) != len(body.MiniBlocks) {
		return process.ErrHeaderBodyMismatch
	}

	var mbHdr data.MiniBlockHeaderHandler
	var miniBlock *block.MiniBlock
	var mbHash []byte
	var err error
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock = body.MiniBlocks[i]
		mbHdr = miniBlockHeaders[i]
		if miniBlock == nil {
			return process.ErrNilMiniBlock
		}

		mbHash, err = core.CalculateHash(bp.marshalizer, bp.hasher, miniBlock)
		if err != nil {
			return err
		}

		err = checkMiniBlockWithMiniBlockHeader(mbHash, mbHdr, miniBlock)
		if err != nil {
			return err
		}

		err = process.CheckIfIndexesAreOutOfBound(mbHdr.GetIndexOfFirstTxProcessed(), mbHdr.GetIndexOfLastTxProcessed(), miniBlock)
		if err != nil {
			return err
		}
		err = checkConstructionStateAndIndexesCorrectness(mbHdr)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkConstructionStateAndIndexesCorrectness(mbh data.MiniBlockHeaderHandler) error {
	if mbh.GetConstructionState() == int32(block.PartialExecuted) && mbh.GetIndexOfLastTxProcessed() == int32(mbh.GetTxCount())-1 {
		return process.ErrIndexDoesNotMatchWithPartialExecutedMiniBlock

	}
	if mbh.GetConstructionState() != int32(block.PartialExecuted) && mbh.GetIndexOfLastTxProcessed() != int32(mbh.GetTxCount())-1 {
		return process.ErrIndexDoesNotMatchWithFullyExecutedMiniBlock
	}

	return nil
}

func (bp *baseProcessor) checkScheduledMiniBlocksValidity(headerHandler data.HeaderHandler) error {
	if !bp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return nil
	}

	scheduledMiniBlocks := bp.scheduledTxsExecutionHandler.GetScheduledMiniBlocks()
	if len(scheduledMiniBlocks) > len(headerHandler.GetMiniBlockHeadersHashes()) {
		log.Debug("baseProcessor.checkScheduledMiniBlocksValidity", "num mbs scheduled", len(scheduledMiniBlocks), "num mbs received", len(headerHandler.GetMiniBlockHeadersHashes()))
		return process.ErrScheduledMiniBlocksMismatch
	}

	for index, scheduledMiniBlock := range scheduledMiniBlocks {
		scheduledMiniBlockHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, scheduledMiniBlock)
		if err != nil {
			return err
		}

		if !bytes.Equal(scheduledMiniBlockHash, headerHandler.GetMiniBlockHeadersHashes()[index]) {
			log.Debug("baseProcessor.checkScheduledMiniBlocksValidity", "index", index, "scheduled mb hash", scheduledMiniBlockHash, "received mb hash", headerHandler.GetMiniBlockHeadersHashes()[index])
			return process.ErrScheduledMiniBlocksMismatch
		}
	}

	return nil
}

func (bp *baseProcessor) requestHeaderByShardAndNonce(shardID uint32, nonce uint64) {
	if shardID == core.MetachainShardId {
		bp.requestHandler.RequestMetaHeaderByNonce(nonce)
	} else {
		bp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
	}
}

func (bp *baseProcessor) cleanExecutionResultsFromTracker(header data.HeaderHandler) error {
	return bp.executionManager.CleanConfirmedExecutionResults(header)
}

func (bp *baseProcessor) cleanupPools(headerHandler data.HeaderHandler) {
	noncesToPrevFinal := bp.getNoncesToFinal(headerHandler) + 1
	bp.cleanupBlockTrackerPools(noncesToPrevFinal)

	highestPrevFinalBlockNonce := bp.forkDetector.GetHighestFinalBlockNonce()
	if highestPrevFinalBlockNonce > 0 {
		highestPrevFinalBlockNonce--
	}

	bp.removeHeadersBehindNonceFromPools(
		true,
		bp.shardCoordinator.SelfId(),
		highestPrevFinalBlockNonce,
	)

	if common.IsFlagEnabledAfterEpochsStartBlock(headerHandler, bp.enableEpochsHandler, common.AndromedaFlag) {
		err := bp.dataPool.Proofs().CleanupProofsBehindNonce(bp.shardCoordinator.SelfId(), highestPrevFinalBlockNonce)
		if err != nil {
			log.Warn("failed to cleanup notarized proofs behind nonce",
				"nonce", noncesToPrevFinal,
				"shardID", bp.shardCoordinator.SelfId(),
				"error", err)
		}
	}

	if bp.shardCoordinator.SelfId() == core.MetachainShardId {
		for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
			bp.cleanupPoolsForCrossShard(shardID, noncesToPrevFinal)
		}
	} else {
		bp.cleanupPoolsForCrossShard(core.MetachainShardId, noncesToPrevFinal)
	}

	for _, executionResult := range headerHandler.GetExecutionResultsHandlers() {
		executionResultHeaderHash := executionResult.GetHeaderHash()
		// cleanup all intra shard miniblocks
		bp.dataPool.MiniBlocks().Remove(executionResultHeaderHash)
		// cleanup all log events
		bp.dataPool.PostProcessTransactions().Remove(common.PrepareLogEventsKey(executionResultHeaderHash))
	}
}

func (bp *baseProcessor) cleanupPoolsForCrossShard(
	shardID uint32,
	noncesToPrevFinal uint64,
) {
	crossNotarizedHeader, _, err := bp.blockTracker.GetCrossNotarizedHeader(shardID, noncesToPrevFinal)
	if err != nil {
		displayCleanupErrorMessage("cleanupPoolsForCrossShard",
			shardID,
			noncesToPrevFinal,
			err)
		return
	}

	crossNotarizedHeaderNonce := common.GetLastExecutionResultNonce(crossNotarizedHeader)

	bp.removeHeadersBehindNonceFromPools(
		false,
		shardID,
		crossNotarizedHeaderNonce,
	)

	if common.IsFlagEnabledAfterEpochsStartBlock(crossNotarizedHeader, bp.enableEpochsHandler, common.AndromedaFlag) {
		err = bp.dataPool.Proofs().CleanupProofsBehindNonce(shardID, noncesToPrevFinal)
		if err != nil {
			log.Warn("failed to cleanup notarized proofs behind nonce",
				"nonce", noncesToPrevFinal,
				"shardID", shardID,
				"error", err)
		}
	}
}

func (bp *baseProcessor) removeHeadersBehindNonceFromPools(
	shouldRemoveBlockBody bool,
	shardId uint32,
	nonce uint64,
) {
	if nonce <= 1 {
		return
	}

	headersPool := bp.dataPool.Headers()
	nonces := headersPool.Nonces(shardId)
	for _, nonceFromCache := range nonces {
		if nonceFromCache >= nonce {
			continue
		}

		if shouldRemoveBlockBody {
			bp.removeBlocksBody(nonceFromCache, shardId)
		}

		headersPool.RemoveHeaderByNonceAndShardId(nonceFromCache, shardId)
	}
}

func (bp *baseProcessor) removeBlocksBody(nonce uint64, shardId uint32) {
	headersPool := bp.dataPool.Headers()
	headers, _, err := headersPool.GetHeadersByNonceAndShardId(nonce, shardId)
	if err != nil {
		return
	}

	for _, header := range headers {
		errNotCritical := bp.removeBlockBodyOfHeader(header)
		if errNotCritical != nil {
			log.Debug("RemoveBlockDataFromPool", "error", errNotCritical.Error())
		}
	}
}

func (bp *baseProcessor) removeBlockBodyOfHeader(headerHandler data.HeaderHandler) error {
	bodyHandler, err := bp.requestBlockBodyHandler.GetProposedAndExecutedMiniBlockHeaders(headerHandler)
	if err != nil {
		return err
	}

	return bp.removeBlockDataFromPools(headerHandler, bodyHandler)
}

func (bp *baseProcessor) removeBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err := bp.txCoordinator.RemoveBlockDataFromPool(body)
	if err != nil {
		return err
	}

	err = bp.blockProcessor.removeStartOfEpochBlockDataFromPools(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	return nil
}

func (bp *baseProcessor) removeTxsFromPools(header data.HeaderHandler, body *block.Body) error {
	newBody, err := bp.getFinalMiniBlocks(header, body)
	if err != nil {
		return err
	}

	rootHashHolder, err := bp.extractRootHashForCleanup(header)
	if err != nil {
		return err
	}

	return bp.txCoordinator.RemoveTxsFromPool(newBody, rootHashHolder)
}

func (bp *baseProcessor) marshalledBodyToBroadcast(body *block.Body) map[uint32][]byte {
	bodies := make(map[uint32]block.MiniBlockSlice)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.SenderShardID != bp.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == bp.shardCoordinator.SelfId() {
			continue
		}
		bodies[miniBlock.ReceiverShardID] = append(bodies[miniBlock.ReceiverShardID], miniBlock)
	}

	marshalledData := make(map[uint32][]byte, len(bodies))
	for shardId, subsetBlockBody := range bodies {
		buff, err := bp.marshalizer.Marshal(&block.Body{MiniBlocks: subsetBlockBody})
		if err != nil {
			log.Error("metaProcessor.MarshalizedDataToBroadcast.Marshal", "error", err.Error())
			continue
		}
		marshalledData[shardId] = buff
	}

	return marshalledData
}

func (bp *baseProcessor) getFinalMiniBlocks(header data.HeaderHandler, body *block.Body) (*block.Body, error) {
	if header.IsHeaderV3() {
		return bp.getFinalMiniBlocksFromExecutionResults(header)
	}

	if !bp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return body, nil
	}

	var miniBlocks block.MiniBlockSlice

	if len(body.MiniBlocks) != len(header.GetMiniBlockHeaderHandlers()) {
		log.Warn("baseProcessor.getFinalMiniBlocks: num of mini blocks and mini blocks headers does not match", "num of mb", len(body.MiniBlocks), "num of mbh", len(header.GetMiniBlockHeaderHandlers()))
		return nil, process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch
	}

	for index, miniBlock := range body.MiniBlocks {
		miniBlockHeader := header.GetMiniBlockHeaderHandlers()[index]
		if !miniBlockHeader.IsFinal() {
			log.Debug("shardProcessor.getFinalMiniBlocks: do not remove from pool / broadcast mini block which is not final", "mb hash", miniBlockHeader.GetHash())
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func (bp *baseProcessor) getFinalMiniBlocksFromExecutionResults(
	header data.HeaderHandler,
) (*block.Body, error) {
	var miniBlocks block.MiniBlockSlice

	baseExecutionResults := header.GetExecutionResultsHandlers()
	if len(baseExecutionResults) == 0 {
		return &block.Body{}, nil
	}

	executedMiniBlocksCache := bp.dataPool.ExecutedMiniBlocks()
	for _, baseExecutionResult := range baseExecutionResults {
		miniBlockHeaderHandlers, err := common.GetMiniBlocksHeaderHandlersFromExecResult(baseExecutionResult)
		if err != nil {
			return nil, err
		}

		for _, miniBlockHeaderHandler := range miniBlockHeaderHandlers {
			mbHash := miniBlockHeaderHandler.GetHash()
			cachedMiniBlock, found := executedMiniBlocksCache.Get(mbHash)
			if !found {
				log.Warn("mini block from execution result not cached after execution",
					"mini block hash", mbHash)
				return nil, process.ErrMissingMiniBlock
			}

			cachedMiniBlockBytes := cachedMiniBlock.([]byte)

			var miniBlock *block.MiniBlock
			err = bp.marshalizer.Unmarshal(&miniBlock, cachedMiniBlockBytes)
			if err != nil {
				return nil, err
			}

			miniBlocks = append(miniBlocks, miniBlock)
		}
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func (bp *baseProcessor) cleanupBlockTrackerPools(noncesToPrevFinal uint64) {
	bp.cleanupBlockTrackerPoolsForShard(bp.shardCoordinator.SelfId(), noncesToPrevFinal)

	if bp.shardCoordinator.SelfId() == core.MetachainShardId {
		for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
			bp.cleanupBlockTrackerPoolsForShard(shardID, noncesToPrevFinal)
		}
	} else {
		bp.cleanupBlockTrackerPoolsForShard(core.MetachainShardId, noncesToPrevFinal)
	}
}

func (bp *baseProcessor) cleanupBlockTrackerPoolsForShard(shardID uint32, noncesToPrevFinal uint64) {
	selfNotarizedHeader, _, errSelfNotarized := bp.blockTracker.GetSelfNotarizedHeader(shardID, noncesToPrevFinal)
	if errSelfNotarized != nil {
		displayCleanupErrorMessage("cleanupBlockTrackerPoolsForShard.GetSelfNotarizedHeader",
			shardID,
			noncesToPrevFinal,
			errSelfNotarized)
		return
	}

	selfNotarizedNonce := common.GetLastExecutionResultNonce(selfNotarizedHeader)

	crossNotarizedNonce := uint64(0)
	if shardID != bp.shardCoordinator.SelfId() {
		crossNotarizedHeader, _, errCrossNotarized := bp.blockTracker.GetCrossNotarizedHeader(shardID, noncesToPrevFinal)
		if errCrossNotarized != nil {
			displayCleanupErrorMessage("cleanupBlockTrackerPoolsForShard.GetCrossNotarizedHeader",
				shardID,
				noncesToPrevFinal,
				errCrossNotarized)
			return
		}

		crossNotarizedNonce = common.GetLastExecutionResultNonce(crossNotarizedHeader)
	}

	bp.blockTracker.CleanupHeadersBehindNonce(
		shardID,
		selfNotarizedNonce,
		crossNotarizedNonce,
	)

	log.Trace("cleanupBlockTrackerPoolsForShard.CleanupHeadersBehindNonce",
		"shard", shardID,
		"self notarized nonce", selfNotarizedNonce,
		"cross notarized nonce", crossNotarizedNonce,
		"nonces to previous final", noncesToPrevFinal)
}

func (bp *baseProcessor) prepareDataForBootStorer(args bootStorerDataArgs) {
	lastCrossNotarizedHeaders := bp.getLastCrossNotarizedHeaders()

	bootData := bootstrapStorage.BootstrapData{
		LastHeader:                 args.headerInfo,
		LastCrossNotarizedHeaders:  lastCrossNotarizedHeaders,
		LastSelfNotarizedHeaders:   args.lastSelfNotarizedHeaders,
		PendingMiniBlocks:          args.pendingMiniBlocks,
		ProcessedMiniBlocks:        args.processedMiniBlocks,
		HighestFinalBlockNonce:     args.highestFinalBlockNonce,
		NodesCoordinatorConfigKey:  args.nodesCoordinatorConfigKey,
		EpochStartTriggerConfigKey: args.epochStartTriggerConfigKey,
	}

	startTime := time.Now()

	err := bp.bootStorer.Put(int64(args.round), bootData)
	if err != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, err,
			"cannot save boot data in storage",
			"err", err)
	}

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveDataForBootStorer", "elapsed time", elapsedTime)
	}
}

func (bp *baseProcessor) getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, bp.shardCoordinator.NumberOfShards()+1)

	for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
		bootstrapHeaderInfo := bp.getLastCrossNotarizedHeadersForShard(shardID)
		if bootstrapHeaderInfo != nil {
			lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
		}
	}

	bootstrapHeaderInfo := bp.getLastCrossNotarizedHeadersForShard(core.MetachainShardId)
	if bootstrapHeaderInfo != nil {
		lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
	}

	if len(lastCrossNotarizedHeaders) == 0 {
		return nil
	}

	return trimSliceBootstrapHeaderInfo(lastCrossNotarizedHeaders)
}

func (bp *baseProcessor) getLastCrossNotarizedHeadersForShard(shardID uint32) *bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bp.blockTracker.GetLastCrossNotarizedHeader(shardID)
	if err != nil {
		log.Warn("getLastCrossNotarizedHeadersForShard",
			"shard", shardID,
			"error", err.Error())
		return nil
	}

	if lastCrossNotarizedHeader.GetNonce() == 0 {
		return nil
	}

	headerInfo := &bootstrapStorage.BootstrapHeaderInfo{
		ShardId: lastCrossNotarizedHeader.GetShardID(),
		Nonce:   lastCrossNotarizedHeader.GetNonce(),
		Hash:    lastCrossNotarizedHeaderHash,
	}

	return headerInfo
}

func (bp *baseProcessor) getLastSelfNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo {
	lastSelfNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, bp.shardCoordinator.NumberOfShards()+1)

	for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
		bootstrapHeaderInfo := bp.getLastSelfNotarizedHeadersForShard(shardID)
		if bootstrapHeaderInfo != nil {
			lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, *bootstrapHeaderInfo)
		}
	}

	bootstrapHeaderInfo := bp.getLastSelfNotarizedHeadersForShard(core.MetachainShardId)
	if bootstrapHeaderInfo != nil {
		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, *bootstrapHeaderInfo)
	}

	if len(lastSelfNotarizedHeaders) == 0 {
		return nil
	}

	return trimSliceBootstrapHeaderInfo(lastSelfNotarizedHeaders)
}

func (bp *baseProcessor) getLastSelfNotarizedHeadersForShard(shardID uint32) *bootstrapStorage.BootstrapHeaderInfo {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := bp.blockTracker.GetLastSelfNotarizedHeader(shardID)
	if err != nil {
		log.Warn("getLastSelfNotarizedHeadersForShard",
			"shard", shardID,
			"error", err.Error())
		return nil
	}

	if lastSelfNotarizedHeader.GetNonce() == 0 {
		return nil
	}

	headerInfo := &bootstrapStorage.BootstrapHeaderInfo{
		ShardId: lastSelfNotarizedHeader.GetShardID(),
		Nonce:   lastSelfNotarizedHeader.GetNonce(),
		Hash:    lastSelfNotarizedHeaderHash,
	}

	return headerInfo
}

func deleteSelfReceiptsMiniBlocks(body *block.Body) *block.Body {
	newBody := &block.Body{}
	for _, mb := range body.MiniBlocks {
		isInShardUnsignedMB := mb.ReceiverShardID == mb.SenderShardID &&
			(mb.Type == block.ReceiptBlock || mb.Type == block.SmartContractResultBlock)
		if isInShardUnsignedMB {
			continue
		}

		newBody.MiniBlocks = append(newBody.MiniBlocks, mb)
	}

	return newBody
}

func (bp *baseProcessor) getNoncesToFinal(headerHandler data.HeaderHandler) uint64 {
	currentBlockNonce := bp.genesisNonce
	if !check.IfNil(headerHandler) {
		currentBlockNonce = headerHandler.GetNonce()
	}

	noncesToFinal := uint64(0)
	finalBlockNonce := bp.getFinalBlockNonce(headerHandler)
	if currentBlockNonce > finalBlockNonce {
		noncesToFinal = currentBlockNonce - finalBlockNonce
	}

	return noncesToFinal
}

func (bp *baseProcessor) getFinalBlockNonce(
	headerHandler data.HeaderHandler,
) uint64 {
	finalBlockNonce := bp.forkDetector.GetHighestFinalBlockNonce()
	if !headerHandler.IsHeaderV3() {
		return finalBlockNonce
	}

	finalHeaderHandler, err := bp.dataPool.Headers().GetHeaderByHash(bp.forkDetector.GetHighestFinalBlockHash())
	if err != nil {
		return finalBlockNonce
	}

	if !finalHeaderHandler.IsHeaderV3() {
		return finalHeaderHandler.GetNonce()
	}

	return common.GetLastExecutionResultNonce(finalHeaderHandler)
}

// DecodeBlockBody method decodes block body from a given byte array
func (bp *baseProcessor) DecodeBlockBody(dta []byte) data.BodyHandler {
	body := &block.Body{}
	if dta == nil {
		return body
	}

	err := bp.marshalizer.Unmarshal(body, dta)
	if err != nil {
		log.Debug("DecodeBlockBody.Unmarshal", "error", err.Error())
		return nil
	}

	return body
}

func (bp *baseProcessor) saveBody(body *block.Body, header data.HeaderHandler, headerHash []byte) {
	startTime := time.Now()
	err := bp.saveProposedTxsToStorage(header, body)
	if err != nil {
		log.Error("saveBody.saveProposedTxsToStorage", "error", err.Error())
	}
	log.Trace("saveBody.SaveTxsToStorage", "time", time.Since(startTime))

	var errNotCritical error
	var marshalizedMiniBlock []byte
	for i := 0; i < len(body.MiniBlocks); i++ {
		marshalizedMiniBlock, errNotCritical = bp.marshalizer.Marshal(body.MiniBlocks[i])
		if errNotCritical != nil {
			log.Warn("saveBody.Marshal", "error", errNotCritical.Error())
			continue
		}

		miniBlockHash := bp.hasher.Compute(string(marshalizedMiniBlock))
		errNotCritical = bp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, marshalizedMiniBlock)
		if errNotCritical != nil {
			logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
				"saveBody.Put -> MiniBlockUnit",
				"err", errNotCritical)
		}
		log.Trace("saveBody.Put -> MiniBlockUnit", "time", time.Since(startTime), "hash", miniBlockHash)
	}

	if !header.IsHeaderV3() {
		errNotCritical = bp.saveReceiptsForHeader(header, headerHash)
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveBody(), error on receiptsRepository.SaveReceipts()",
			"err", errNotCritical)
	}

	bp.scheduledTxsExecutionHandler.SaveStateIfNeeded(headerHash)

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveBody", "elapsed time", elapsedTime)
	}
}

func (bp *baseProcessor) saveProposedTxsToStorage(header data.HeaderHandler, body *block.Body) error {
	if !header.IsHeaderV3() {
		bp.txCoordinator.SaveTxsToStorage(body)
		return nil
	}

	separatedBodies := process.SeparateBodyByType(body)
	for blockType, blockBody := range separatedBodies {
		dataPool, err := bp.getDataPoolByBlockType(blockType)
		if err != nil {
			return err
		}

		unit, err := getStorageForProposedTxsFromBlockType(blockType)
		if err != nil {
			return err
		}

		storer, err := bp.store.GetStorer(unit)
		if err != nil {
			return err
		}

		for i := 0; i < len(blockBody.MiniBlocks); i++ {
			miniBlock := blockBody.MiniBlocks[i]
			err = bp.saveTxsToStorage(dataPool, storer, miniBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (bp *baseProcessor) saveTxsToStorage(dataPool dataRetriever.ShardedDataCacherNotifier, storer storage.Storer, miniBlock *block.MiniBlock) error {
	txHashes := miniBlock.TxHashes
	senderShardID := miniBlock.SenderShardID
	receiverShardID := miniBlock.ReceiverShardID
	method := process.SearchMethodPeekWithFallbackSearchFirst

	for _, txHash := range txHashes {
		tx, err := process.GetTransactionHandlerFromPool(senderShardID, receiverShardID, txHash, dataPool, method)
		if err != nil {
			return err
		}

		marshalledTx, err := bp.marshalizer.Marshal(tx)
		if err != nil {
			return err
		}

		err = storer.Put(txHash, marshalledTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bp *baseProcessor) getDataPoolByBlockType(blockType block.Type) (dataRetriever.ShardedDataCacherNotifier, error) {
	switch blockType {
	case block.TxBlock, block.InvalidBlock:
		return bp.dataPool.Transactions(), nil
	case block.SmartContractResultBlock:
		return bp.dataPool.UnsignedTransactions(), nil
	case block.RewardsBlock:
		return bp.dataPool.RewardTransactions(), nil
	case block.PeerBlock:
		return bp.dataPool.ValidatorsInfo(), nil
	default:
		return nil, fmt.Errorf("unsupported block type for dataPool: %d", blockType)
	}
}

func getStorageForProposedTxsFromBlockType(blockType block.Type) (dataRetriever.UnitType, error) {
	switch blockType {
	case block.TxBlock, block.InvalidBlock:
		return dataRetriever.TransactionUnit, nil
	case block.SmartContractResultBlock:
		return dataRetriever.UnsignedTransactionUnit, nil
	case block.ReceiptBlock:
		return dataRetriever.ReceiptsUnit, nil
	case block.RewardsBlock:
		return dataRetriever.RewardTransactionUnit, nil
	case block.PeerBlock:
		return dataRetriever.UnsignedTransactionUnit, nil
	}
	return 0, process.ErrInvalidBlockType
}

func (bp *baseProcessor) saveShardHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	startTime := time.Now()

	nonceToByteSlice := bp.uint64Converter.ToByteSlice(header.GetNonce())
	hdrNonceHashDataUnit := dataRetriever.GetHdrNonceHashDataUnit(header.GetShardID())

	errNotCritical := bp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			fmt.Sprintf("saveHeader.Put -> ShardHdrNonceHashDataUnit_%d", header.GetShardID()),
			"err", errNotCritical)
	}

	errNotCritical = bp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, marshalizedHeader)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveHeader.Put -> BlockHeaderUnit",
			"err", errNotCritical)
	}

	bp.saveProof(headerHash, header)

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveShardHeader", "elapsed time", elapsedTime)
	}
}

func (bp *baseProcessor) saveMetaHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	startTime := time.Now()

	nonceToByteSlice := bp.uint64Converter.ToByteSlice(header.GetNonce())

	errNotCritical := bp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveMetaHeader.Put -> MetaHdrNonceHashDataUnit",
			"err", errNotCritical)
	}

	errNotCritical = bp.store.Put(dataRetriever.MetaBlockUnit, headerHash, marshalizedHeader)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveMetaHeader.Put -> MetaBlockUnit",
			"err", errNotCritical)
	}

	bp.saveProof(headerHash, header)

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveMetaHeader", "elapsed time", elapsedTime)
	}
}

func (bp *baseProcessor) saveProof(
	hash []byte,
	header data.HeaderHandler,
) {
	if !common.IsProofsFlagEnabledForHeader(bp.enableEpochsHandler, header) {
		return
	}

	proof, err := bp.proofsPool.GetProof(header.GetShardID(), hash)
	if err != nil {
		log.Error("could not find proof for header",
			"hash", hex.EncodeToString(hash),
			"shard", header.GetShardID(),
		)
		return
	}
	marshalledProof, errNotCritical := bp.marshalizer.Marshal(proof)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveProof.Marshal proof",
			"err", errNotCritical)
		return
	}

	errNotCritical = bp.store.Put(dataRetriever.ProofsUnit, proof.GetHeaderHash(), marshalledProof)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveProof.Put -> ProofsUnit",
			"err", errNotCritical)
	}

	log.Trace("saved proof to storage", "hash", hash)
}

func getLastSelfNotarizedHeaderByItself(chainHandler data.ChainHandler) (data.HeaderHandler, []byte) {
	currentHeader := chainHandler.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return chainHandler.GetGenesisHeader(), chainHandler.GetGenesisHeaderHash()
	}

	currentBlockHash := chainHandler.GetCurrentBlockHeaderHash()

	return currentHeader, currentBlockHash
}

func (bp *baseProcessor) setFinalizedHeaderHashInIndexer(hdrHash []byte) {
	log.Debug("baseProcessor.setFinalizedHeaderHashInIndexer", "finalized header hash", hdrHash)

	bp.outportHandler.FinalizedBlock(&outportcore.FinalizedBlock{ShardID: bp.shardCoordinator.SelfId(), HeaderHash: hdrHash})
}

func (bp *baseProcessor) updateStateStorage(
	finalHeaderNonce uint64,
	currRootHash []byte,
	prevRootHash []byte,
	accounts state.AccountsAdapter,
) {
	if !accounts.IsPruningEnabled() {
		return
	}

	if bytes.Equal(prevRootHash, currRootHash) {
		return
	}

	accounts.CancelPrune(prevRootHash, state.NewRoot)
	accounts.PruneTrie(prevRootHash, state.OldRoot, bp.getPruningHandler(finalHeaderNonce))
}

// RevertCurrentBlock reverts the current block for cleanup failed process
func (bp *baseProcessor) RevertCurrentBlock(headerHandler data.HeaderHandler) {
	bp.revertAccountState()
	bp.revertScheduledInfo()
	bp.revertCurrentBlockV3(headerHandler)
}

func (bp *baseProcessor) revertCurrentBlockV3(headerHandler data.HeaderHandler) {
	if check.IfNil(headerHandler) || !headerHandler.IsHeaderV3() {
		return
	}

	headerNonce := headerHandler.GetNonce()
	err := bp.executionManager.RemoveAtNonceAndHigher(headerNonce)
	if err != nil {
		log.Debug("baseProcessor.revertCurrentBlockV3", "err", err)
		lastExecResult, errGet := common.GetLastBaseExecutionResultHandler(headerHandler)
		if errGet != nil {
			log.Error("baseProcessor.revertCurrentBlockV3.GetLastBaseExecutionResultHandler", "err", errGet)
			return
		}
		errReset := bp.executionManager.ResetAndResumeExecution(lastExecResult)
		if errReset != nil {
			log.Debug("baseProcessor.revertCurrentBlockV3.ResetAndResumeExecution", "err", errReset)
		}
	}
}

func (bp *baseProcessor) revertAccountState() {
	for key := range bp.accountsDB {
		err := bp.accountsDB[key].RevertToSnapshot(0)
		if err != nil {
			log.Debug("RevertToSnapshot", "error", err.Error())
		}
	}
}

func (bp *baseProcessor) revertScheduledInfo() {
	header, headerHash := bp.getLastCommittedHeaderAndHash()
	if header.IsHeaderV3() {
		// v3 headers don't have scheduled info
		return
	}

	err := bp.scheduledTxsExecutionHandler.RollBackToBlock(headerHash)
	if err != nil {
		log.Trace("baseProcessor.revertScheduledInfo", "error", err.Error())
		scheduledInfo := &process.ScheduledInfo{
			RootHash:        header.GetRootHash(),
			IntermediateTxs: make(map[block.Type][]data.TransactionHandler),
			GasAndFees:      process.GetZeroGasAndFees(),
			MiniBlocks:      make(block.MiniBlockSlice, 0),
		}
		bp.scheduledTxsExecutionHandler.SetScheduledInfo(scheduledInfo)
	}
}

func (bp *baseProcessor) getLastCommittedHeaderAndHash() (data.HeaderHandler, []byte) {
	headerHandler := bp.blockChain.GetCurrentBlockHeader()
	headerHash := bp.blockChain.GetCurrentBlockHeaderHash()
	if check.IfNil(headerHandler) {
		headerHandler = bp.blockChain.GetGenesisHeader()
		headerHash = bp.blockChain.GetGenesisHeaderHash()
	}

	return headerHandler, headerHash
}

// GetAccountsDBSnapshot returns the account snapshot
func (bp *baseProcessor) GetAccountsDBSnapshot() map[state.AccountsDbIdentifier]int {
	snapshots := make(map[state.AccountsDbIdentifier]int)
	for key := range bp.accountsDB {
		snapshots[key] = bp.accountsDB[key].JournalLen()
	}

	return snapshots
}

// RevertAccountsDBToSnapshot reverts the accountsDB to the given snapshot
func (bp *baseProcessor) RevertAccountsDBToSnapshot(accountsSnapshot map[state.AccountsDbIdentifier]int) {
	for key := range bp.accountsDB {
		err := bp.accountsDB[key].RevertToSnapshot(accountsSnapshot[key])
		if err != nil {
			log.Debug("RevertAccountsDBToSnapshot", "error", err.Error())
		}
	}
}

func (bp *baseProcessor) commitState(headerHandler data.HeaderHandler) error {
	startTime := time.Now()
	defer func() {
		elapsedTime := time.Since(startTime)
		log.Debug("elapsed time to commit accounts state",
			"time [s]", elapsedTime,
			"header nonce", headerHandler.GetNonce(),
		)
	}()

	if headerHandler.IsStartOfEpochBlock() {
		return bp.commitInLastEpoch(headerHandler.GetEpoch())
	}

	return bp.commit()
}

func (bp *baseProcessor) commitInLastEpoch(currentEpoch uint32) error {
	lastEpoch := uint32(0)
	if currentEpoch > 0 {
		lastEpoch = currentEpoch - 1
	}

	return bp.commitInEpoch(currentEpoch, lastEpoch)
}

func (bp *baseProcessor) commit() error {
	for key := range bp.accountsDB {
		_, err := bp.accountsDB[key].Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (bp *baseProcessor) commitInEpoch(currentEpoch uint32, epochToCommit uint32) error {
	for key := range bp.accountsDB {
		_, err := bp.accountsDB[key].CommitInEpoch(currentEpoch, epochToCommit)
		if err != nil {
			return err
		}
	}

	return nil
}

// PruneStateOnRollback recreates the state tries to the root hashes indicated by the provided headers
func (bp *baseProcessor) PruneStateOnRollback(currHeader data.HeaderHandler, currHeaderHash []byte, prevHeader data.HeaderHandler, prevHeaderHash []byte) {
	for key := range bp.accountsDB {
		if !bp.accountsDB[key].IsPruningEnabled() {
			continue
		}

		rootHash, prevRootHash := bp.getRootHashes(currHeader, prevHeader, key)
		if key == state.UserAccountsState {
			scheduledRootHash, err := bp.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(currHeaderHash)
			if err == nil {
				rootHash = scheduledRootHash
			}

			scheduledPrevRootHash, err := bp.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(prevHeaderHash)
			if err == nil {
				prevRootHash = scheduledPrevRootHash
			}

			var prevStartScheduledRootHash []byte
			if prevHeader.GetAdditionalData() != nil && prevHeader.GetAdditionalData().GetScheduledRootHash() != nil {
				prevStartScheduledRootHash = prevHeader.GetAdditionalData().GetScheduledRootHash()
				if bytes.Equal(prevStartScheduledRootHash, prevRootHash) {
					bp.accountsDB[key].CancelPrune(prevStartScheduledRootHash, state.OldRoot)
				}
			}
		}

		if bytes.Equal(rootHash, prevRootHash) {
			continue
		}

		bp.accountsDB[key].CancelPrune(prevRootHash, state.OldRoot)
		bp.accountsDB[key].PruneTrie(rootHash, state.NewRoot, bp.getPruningHandler(currHeader.GetNonce()))
	}
}

func (bp *baseProcessor) getPruningHandler(finalHeaderNonce uint64) state.PruningHandler {
	if finalHeaderNonce-bp.lastRestartNonce <= uint64(bp.pruningDelay) {
		log.Debug("will skip pruning",
			"finalHeaderNonce", finalHeaderNonce,
			"last restart nonce", bp.lastRestartNonce,
			"num blocks for pruning delay", bp.pruningDelay,
		)
		return state.NewPruningHandler(state.DisableDataRemoval)
	}

	return state.NewPruningHandler(state.EnableDataRemoval)
}

func (bp *baseProcessor) getRootHashes(currHeader data.HeaderHandler, prevHeader data.HeaderHandler, identifier state.AccountsDbIdentifier) ([]byte, []byte) {
	switch identifier {
	case state.UserAccountsState:
		return currHeader.GetRootHash(), prevHeader.GetRootHash()
	case state.PeerAccountsState:
		currMetaHeader, ok := currHeader.(data.MetaHeaderHandler)
		if !ok {
			return []byte{}, []byte{}
		}
		prevMetaHeader, ok := prevHeader.(data.MetaHeaderHandler)
		if !ok {
			return []byte{}, []byte{}
		}
		return currMetaHeader.GetValidatorStatsRootHash(), prevMetaHeader.GetValidatorStatsRootHash()
	default:
		return []byte{}, []byte{}
	}
}

func (bp *baseProcessor) displayMiniBlocksPool() {
	miniBlocksPool := bp.dataPool.MiniBlocks()

	for _, hash := range miniBlocksPool.Keys() {
		value, ok := miniBlocksPool.Get(hash)
		if !ok {
			log.Debug("displayMiniBlocksPool: mini block not found", "hash", logger.DisplayByteSlice(hash))
			continue
		}

		miniBlock, ok := value.(*block.MiniBlock)
		if !ok {
			log.Debug("displayMiniBlocksPool: wrong type assertion", "hash", logger.DisplayByteSlice(hash))
			continue
		}

		log.Trace("mini block in pool",
			"hash", logger.DisplayByteSlice(hash),
			"type", miniBlock.Type,
			"sender", miniBlock.SenderShardID,
			"receiver", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes))
	}
}

// trimSliceBootstrapHeaderInfo creates a copy of the provided slice without the excess capacity
func trimSliceBootstrapHeaderInfo(in []bootstrapStorage.BootstrapHeaderInfo) []bootstrapStorage.BootstrapHeaderInfo {
	if len(in) == 0 {
		return []bootstrapStorage.BootstrapHeaderInfo{}
	}
	ret := make([]bootstrapStorage.BootstrapHeaderInfo, len(in))
	copy(ret, in)
	return ret
}

func (bp *baseProcessor) restoreBlockBody(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) {
	if check.IfNil(bodyHandler) {
		log.Debug("restoreMiniblocks nil bodyHandler")
		return
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		log.Debug("restoreMiniblocks wrong type assertion for bodyHandler")
		return
	}

	_, errNotCritical := bp.txCoordinator.RestoreBlockDataFromStorage(body)
	if errNotCritical != nil {
		log.Debug("restoreBlockBody RestoreBlockDataFromStorage", "error", errNotCritical.Error())
	}

	go bp.txCounter.headerReverted(headerHandler)
}

// RemoveHeaderFromPool removes the header from the pool
func (bp *baseProcessor) RemoveHeaderFromPool(headerHash []byte) {
	headersPool := bp.dataPool.Headers()
	headersPool.RemoveHeaderByHash(headerHash)
}

// RestoreBlockBodyIntoPools restores the block body into associated pools
func (bp *baseProcessor) RestoreBlockBodyIntoPools(bodyHandler data.BodyHandler) error {
	if check.IfNil(bodyHandler) {
		return process.ErrNilBlockBody
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	_, err := bp.txCoordinator.RestoreBlockDataFromStorage(body)
	if err != nil {
		return err
	}

	return nil
}

func (bp *baseProcessor) recordBlockInHistory(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) {
	scrResultsFromPool := bp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	receiptsFromPool := bp.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock)
	logs := bp.txCoordinator.GetAllCurrentLogs()
	intraMiniBlocks := bp.txCoordinator.GetCreatedInShardMiniBlocks()

	err := bp.historyRepo.RecordBlock(blockHeaderHash, blockHeader, blockBody, scrResultsFromPool, receiptsFromPool, intraMiniBlocks, logs)
	if err != nil {
		logLevel := logger.LogError
		if core.IsClosingError(err) {
			logLevel = logger.LogDebug
		}
		log.Log(logLevel, "historyRepo.RecordBlock()", "blockHeaderHash", blockHeaderHash, "error", err.Error())
	}
}

func (bp *baseProcessor) addHeaderIntoTrackerPool(nonce uint64, shardID uint32) {
	headersPool := bp.dataPool.Headers()
	headers, hashes, err := headersPool.GetHeadersByNonceAndShardId(nonce, shardID)
	if err != nil {
		log.Trace("baseProcessor.addHeaderIntoTrackerPool", "error", err.Error())
		return
	}

	for i := 0; i < len(headers); i++ {
		bp.blockTracker.AddTrackedHeader(headers[i], hashes[i])
	}
}

func (bp *baseProcessor) commitTrieEpochRootHashIfNeeded(metaBlock data.MetaHeaderHandler, rootHash []byte) error {
	trieEpochRootHashStorageUnit, err := bp.store.GetStorer(dataRetriever.TrieEpochRootHashUnit)
	if err != nil {
		return err
	}

	if check.IfNil(trieEpochRootHashStorageUnit) {
		return nil
	}
	_, isStorerDisabled := trieEpochRootHashStorageUnit.(*storageunit.NilStorer)
	if isStorerDisabled {
		return nil
	}

	userAccountsDb := bp.accountsDB[state.UserAccountsState]
	if userAccountsDb == nil {
		return fmt.Errorf("%w for user accounts state", process.ErrNilAccountsAdapter)
	}

	epochBytes := bp.uint64Converter.ToByteSlice(uint64(metaBlock.GetEpoch()))

	err = trieEpochRootHashStorageUnit.Put(epochBytes, rootHash)
	if err != nil {
		return err
	}

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = userAccountsDb.GetAllLeaves(iteratorChannels, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
	if err != nil {
		return err
	}

	processDataTries := bp.processDataTriesOnCommitEpoch
	balanceSum := big.NewInt(0)
	numAccountLeaves := 0
	numAccountsWithDataTrie := 0
	numCodeLeaves := 0
	totalSizeAccounts := 0
	totalSizeAccountsDataTries := 0
	totalSizeCodeLeaves := 0

	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:                 bp.hasher,
		Marshaller:             bp.marshalizer,
		EnableEpochsHandler:    bp.enableEpochsHandler,
		StateAccessesCollector: bp.stateAccessesCollector,
	}
	accountCreator, err := factory.NewAccountCreator(argsAccCreator)
	if err != nil {
		return err
	}

	for leaf := range iteratorChannels.LeavesChan {
		userAccount, errUnmarshal := bp.unmarshalUserAccount(accountCreator, leaf.Key(), leaf.Value())
		if errUnmarshal != nil {
			numCodeLeaves++
			totalSizeCodeLeaves += len(leaf.Value())
			log.Trace("cannot unmarshal user account. it may be a code leaf", "error", errUnmarshal)
			continue
		}

		if processDataTries {
			rh := userAccount.GetRootHash()
			if len(rh) != 0 {
				dataTrie := &common.TrieIteratorChannels{
					LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
					ErrChan:    errChan.NewErrChanWrapper(),
				}
				errDataTrieGet := userAccountsDb.GetAllLeaves(dataTrie, context.Background(), rh, parsers.NewMainTrieLeafParser())
				if errDataTrieGet != nil {
					continue
				}

				currentSize := 0
				for lf := range dataTrie.LeavesChan {
					currentSize += len(lf.Value())
				}

				err = dataTrie.ErrChan.ReadFromChanNonBlocking()
				if err != nil {
					return err
				}

				totalSizeAccountsDataTries += currentSize
				numAccountsWithDataTrie++
			}
		}

		numAccountLeaves++
		totalSizeAccounts += len(leaf.Value())

		balanceSum.Add(balanceSum, userAccount.GetBalance())
	}

	err = iteratorChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return err
	}

	totalSizeAccounts += totalSizeAccountsDataTries

	stats := []interface{}{
		"shard", bp.shardCoordinator.SelfId(),
		"epoch", metaBlock.GetEpoch(),
		"sum", balanceSum.String(),
		"processDataTries", processDataTries,
		"numCodeLeaves", numCodeLeaves,
		"totalSizeCodeLeaves", totalSizeCodeLeaves,
		"numAccountLeaves", numAccountLeaves,
		"totalSizeAccountsLeaves", totalSizeAccounts,
	}

	if processDataTries {
		stats = append(stats, []interface{}{
			"from which numAccountsWithDataTrie", numAccountsWithDataTrie,
			"from which totalSizeAccountsDataTries", totalSizeAccountsDataTries}...)
	}

	log.Debug("sum of addresses in shard at epoch start", stats...)

	return nil
}

func (bp *baseProcessor) unmarshalUserAccount(
	accountCreator state.AccountFactory,
	address []byte,
	userAccountsBytes []byte,
) (state.UserAccountHandler, error) {
	account, err := accountCreator.CreateAccount(address)
	if err != nil {
		return nil, err
	}
	err = bp.marshalizer.Unmarshal(account, userAccountsBytes)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAccount, nil
}

// Close - closes all underlying components
func (bp *baseProcessor) Close() error {
	var err1, err2, err3 error
	if !check.IfNil(bp.vmContainer) {
		err1 = bp.vmContainer.Close()
	}
	if !check.IfNil(bp.vmContainerFactory) {
		err2 = bp.vmContainerFactory.Close()
	}
	err3 = bp.processDebugger.Close()
	if err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("vmContainer close error: %v, vmContainerFactory close error: %v, processDebugger close: %v",
			err1, err2, err3)
	}

	return nil
}

// ProcessScheduledBlock processes a scheduled block
func (bp *baseProcessor) ProcessScheduledBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler, haveTime func() time.Duration) error {
	var err error
	bp.processStatusHandler.SetBusy("baseProcessor.ProcessScheduledBlock")
	defer func() {
		if err != nil {
			bp.RevertCurrentBlock(headerHandler)
		}
		bp.processStatusHandler.SetIdle()
	}()

	scheduledMiniBlocksFromMe, err := getScheduledMiniBlocksFromMe(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	bp.scheduledTxsExecutionHandler.AddScheduledMiniBlocks(scheduledMiniBlocksFromMe)

	normalProcessingGasAndFees := bp.getGasAndFees()

	startTime := time.Now()
	err = bp.scheduledTxsExecutionHandler.ExecuteAll(haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to execute all scheduled transactions",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	rootHash, err := bp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	_ = bp.txCoordinator.CreatePostProcessMiniBlocks()

	finalProcessingGasAndFees := bp.getGasAndFeesWithScheduled()

	scheduledProcessingGasAndFees := gasAndFeesDelta(normalProcessingGasAndFees, finalProcessingGasAndFees)
	bp.scheduledTxsExecutionHandler.SetScheduledRootHash(rootHash)
	bp.scheduledTxsExecutionHandler.SetScheduledGasAndFees(scheduledProcessingGasAndFees)

	return nil
}

func getScheduledMiniBlocksFromMe(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) (block.MiniBlockSlice, error) {
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	if len(body.MiniBlocks) != len(headerHandler.GetMiniBlockHeaderHandlers()) {
		log.Warn("getScheduledMiniBlocksFromMe: num of mini blocks and mini blocks headers does not match", "num of mb", len(body.MiniBlocks), "num of mbh", len(headerHandler.GetMiniBlockHeaderHandlers()))
		return nil, process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	for index, miniBlock := range body.MiniBlocks {
		miniBlockHeader := headerHandler.GetMiniBlockHeaderHandlers()[index]
		isScheduledMiniBlockFromMe := miniBlockHeader.GetSenderShardID() == headerHandler.GetShardID() && miniBlockHeader.GetProcessingType() == int32(block.Scheduled)
		if isScheduledMiniBlockFromMe {
			miniBlocks = append(miniBlocks, miniBlock)
		}

	}

	return miniBlocks, nil
}

func (bp *baseProcessor) getGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: bp.feeHandler.GetAccumulatedFees(),
		DeveloperFees:   bp.feeHandler.GetDeveloperFees(),
		GasProvided:     bp.gasConsumedProvider.TotalGasProvided(),
		GasPenalized:    bp.gasConsumedProvider.TotalGasPenalized(),
		GasRefunded:     bp.gasConsumedProvider.TotalGasRefunded(),
	}
}

func (bp *baseProcessor) getGasAndFeesWithScheduled() scheduled.GasAndFees {
	gasAndFees := bp.getGasAndFees()
	gasAndFees.GasProvided = bp.gasConsumedProvider.TotalGasProvidedWithScheduled()
	return gasAndFees
}

func gasAndFeesDelta(initialGasAndFees, finalGasAndFees scheduled.GasAndFees) scheduled.GasAndFees {
	zero := big.NewInt(0)
	result := process.GetZeroGasAndFees()

	deltaAccumulatedFees := big.NewInt(0).Sub(finalGasAndFees.AccumulatedFees, initialGasAndFees.AccumulatedFees)
	if deltaAccumulatedFees.Cmp(zero) < 0 {
		log.Error("gasAndFeesDelta",
			"initial accumulatedFees", initialGasAndFees.AccumulatedFees.String(),
			"final accumulatedFees", finalGasAndFees.AccumulatedFees.String(),
			"error", process.ErrNegativeValue)
		return result
	}

	deltaDevFees := big.NewInt(0).Sub(finalGasAndFees.DeveloperFees, initialGasAndFees.DeveloperFees)
	if deltaDevFees.Cmp(zero) < 0 {
		log.Error("gasAndFeesDelta",
			"initial devFees", initialGasAndFees.DeveloperFees.String(),
			"final devFees", finalGasAndFees.DeveloperFees.String(),
			"error", process.ErrNegativeValue)
		return result
	}

	deltaGasProvided := int64(finalGasAndFees.GasProvided) - int64(initialGasAndFees.GasProvided)
	if deltaGasProvided < 0 {
		log.Error("gasAndFeesDelta",
			"initial gasProvided", initialGasAndFees.GasProvided,
			"final gasProvided", finalGasAndFees.GasProvided,
			"error", process.ErrNegativeValue)
		return result
	}

	deltaGasPenalized := int64(finalGasAndFees.GasPenalized) - int64(initialGasAndFees.GasPenalized)
	if deltaGasPenalized < 0 {
		log.Error("gasAndFeesDelta",
			"initial gasPenalized", initialGasAndFees.GasPenalized,
			"final gasPenalized", finalGasAndFees.GasPenalized,
			"error", process.ErrNegativeValue)
		return result
	}

	deltaGasRefunded := int64(finalGasAndFees.GasRefunded) - int64(initialGasAndFees.GasRefunded)
	if deltaGasRefunded < 0 {
		log.Error("gasAndFeesDelta",
			"initial gasRefunded", initialGasAndFees.GasRefunded,
			"final gasRefunded", finalGasAndFees.GasRefunded,
			"error", process.ErrNegativeValue)
		return result
	}

	return scheduled.GasAndFees{
		AccumulatedFees: deltaAccumulatedFees,
		DeveloperFees:   deltaDevFees,
		GasProvided:     uint64(deltaGasProvided),
		GasPenalized:    uint64(deltaGasPenalized),
		GasRefunded:     uint64(deltaGasRefunded),
	}
}

func (bp *baseProcessor) getIndexOfFirstMiniBlockToBeExecuted(header data.HeaderHandler) int {
	if !bp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return 0
	}

	for index, miniBlockHeaderHandler := range header.GetMiniBlockHeaderHandlers() {
		if miniBlockHeaderHandler.GetProcessingType() == int32(block.Processed) {
			log.Debug("baseProcessor.getIndexOfFirstMiniBlockToBeExecuted: mini block is already executed",
				"mb hash", miniBlockHeaderHandler.GetHash(),
				"mb index", index)
			continue
		}

		return index
	}

	return len(header.GetMiniBlockHeaderHandlers())
}

func displayCleanupErrorMessage(message string, shardID uint32, noncesToPrevFinal uint64, err error) {
	// 2 blocks on shard + 2 blocks on meta + 1 block to previous final
	maxNoncesToPrevFinalWithoutWarn := uint64(process.BlockFinality+1)*2 + 1
	level := logger.LogWarning
	if noncesToPrevFinal <= maxNoncesToPrevFinalWithoutWarn {
		level = logger.LogDebug
	}

	log.Log(level, message,
		"shard", shardID,
		"nonces to previous final", noncesToPrevFinal,
		"error", err.Error())
}

// SetProcessDebugger sets the process debugger associated to this block processor
func (bp *baseProcessor) SetProcessDebugger(debugger process.Debugger) error {
	if check.IfNil(debugger) {
		return process.ErrNilProcessDebugger
	}

	bp.mutProcessDebugger.Lock()
	bp.processDebugger = debugger
	bp.mutProcessDebugger.Unlock()

	return nil
}

func (bp *baseProcessor) updateLastCommittedInDebugger(round uint64) {
	bp.mutProcessDebugger.RLock()
	bp.processDebugger.SetLastCommittedBlockRound(round)
	bp.mutProcessDebugger.RUnlock()
}

func createDisabledProcessDebugger() (process.Debugger, error) {
	configs := config.ProcessDebugConfig{
		Enabled: false,
	}

	return debugFactory.CreateProcessDebugger(configs)
}

// NonceOfFirstCommittedBlock returns the first committed block's nonce. The optional Uint64 will contain a-not-set value
// if no block was committed by the node
func (bp *baseProcessor) NonceOfFirstCommittedBlock() core.OptionalUint64 {
	bp.mutNonceOfFirstCommittedBlock.RLock()
	defer bp.mutNonceOfFirstCommittedBlock.RUnlock()

	return bp.nonceOfFirstCommittedBlock
}

func (bp *baseProcessor) setNonceOfFirstCommittedBlock(nonce uint64) {
	bp.mutNonceOfFirstCommittedBlock.Lock()
	defer bp.mutNonceOfFirstCommittedBlock.Unlock()

	if bp.nonceOfFirstCommittedBlock.HasValue {
		return
	}

	bp.nonceOfFirstCommittedBlock.HasValue = true
	bp.nonceOfFirstCommittedBlock.Value = nonce
}

// OnProposedBlock calls the OnProposedBlock from transactions pool
func (bp *baseProcessor) OnProposedBlock(
	proposedBody data.BodyHandler,
	proposedHeader data.HeaderHandler,
	proposedHash []byte,
) error {
	// this should be removed once OnProposedBlock accepts bodyHandler
	proposedBodyPtr, ok := proposedBody.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	// TODO: call SetRootHashIfNeeded for accountsProposal which should recreate the trie if needed for
	accountsProvider, err := state.NewAccountsEphemeralProvider(bp.accountsProposal)
	if err != nil {
		return err
	}

	lastCommittedHeader, err := bp.dataPool.Headers().GetHeaderByHash(proposedHeader.GetPrevHash())
	if err != nil {
		return err
	}

	lastExecResHandler, err := common.GetOrCreateLastExecutionResultForPrevHeader(lastCommittedHeader, proposedHeader.GetPrevHash())
	if err != nil {
		return err
	}

	return bp.dataPool.Transactions().OnProposedBlock(proposedHash, proposedBodyPtr, proposedHeader, accountsProvider, lastExecResHandler.GetHeaderHash())
}

func (bp *baseProcessor) onExecutedBlock(header data.HeaderHandler, rootHash []byte) error {
	bp.appStatusHandler.SetUInt64Value(common.MetricNumTrackedBlocks, bp.dataPool.Transactions().GetNumTrackedBlocks())
	bp.appStatusHandler.SetUInt64Value(common.MetricNumTrackedAccounts, bp.dataPool.Transactions().GetNumTrackedAccounts())

	err := bp.dataPool.Transactions().OnExecutedBlock(header, rootHash)
	if err != nil {
		log.Error("baseProcessor.onExecutedBlock", "err", err)
		return err
	}

	return nil
}

func (bp *baseProcessor) recreateTrieIfNeeded() error {
	rootHash := bp.blockChain.GetCurrentBlockRootHash()
	if len(rootHash) == 0 {
		genesisBlock := bp.blockChain.GetGenesisHeader()
		rootHash = genesisBlock.GetRootHash()
	}

	rh := holders.NewDefaultRootHashesHolder(rootHash)
	err := bp.accountsProposal.RecreateTrieIfNeeded(rh)
	if err != nil {
		log.Error("baseProcessor.recreateTrieIfNeeded", "err", err)
		return err
	}

	return nil
}

func (bp *baseProcessor) extractRootHashForCleanup(header data.HeaderHandler) (common.RootHashHolder, error) {
	if header.IsHeaderV3() {
		latestExecutionResult, err := common.GetLastBaseExecutionResultHandler(header)
		if err != nil {
			return nil, err
		}

		return holders.NewDefaultRootHashesHolder(latestExecutionResult.GetRootHash()), nil
	}

	additionalData := header.GetAdditionalData()
	if additionalData == nil {
		rootHash := bp.blockChain.GetCurrentBlockRootHash()
		return holders.NewDefaultRootHashesHolder(rootHash), nil
	}

	return holders.NewDefaultRootHashesHolder(additionalData.GetScheduledRootHash()), nil
}

func (bp *baseProcessor) checkSentSignaturesAtCommitTime(header data.HeaderHandler) error {
	_, validatorsGroup, err := headerCheck.ComputeConsensusGroup(header, bp.nodesCoordinator)
	if err != nil {
		return err
	}

	consensusGroup := make([]string, 0, len(validatorsGroup))
	for _, validator := range validatorsGroup {
		consensusGroup = append(consensusGroup, string(validator.PubKey()))
	}

	signers := headerCheck.ComputeSignersPublicKeys(consensusGroup, header.GetPubKeysBitmap())

	for _, signer := range signers {
		bp.sentSignaturesTracker.ResetCountersForManagedBlockSigner([]byte(signer))
	}

	return nil
}

func (bp *baseProcessor) getHeaderHash(header data.HeaderHandler) ([]byte, error) {
	marshalledHeader, errMarshal := bp.marshalizer.Marshal(header)
	if errMarshal != nil {
		return nil, errMarshal
	}

	return bp.hasher.Compute(string(marshalledHeader)), nil
}

func (bp *baseProcessor) computeOwnShardStuckIfNeeded(header data.HeaderHandler) error {
	if !header.IsHeaderV3() {
		return nil
	}

	lastExecResultsHandler, err := common.GetLastBaseExecutionResultHandler(header)
	if err != nil {
		return err
	}

	bp.blockTracker.ComputeOwnShardStuck(lastExecResultsHandler, header.GetNonce())
	return nil
}

func (bp *baseProcessor) updateGasConsumptionLimitsIfNeeded() {
	if !bp.blockTracker.IsOwnShardStuck() {
		bp.gasComputation.ResetIncomingLimit()
		bp.gasComputation.ResetOutgoingLimit()

		return
	}

	// shard is stuck, zeroing the limits
	bp.gasComputation.ZeroIncomingLimit()
	bp.gasComputation.ZeroOutgoingLimit()
}

func (bp *baseProcessor) getMaxRoundsWithoutBlockReceived(round uint64) uint64 {
	maxRoundsWithoutNewBlockReceived := bp.processConfigsHandler.GetMaxRoundsWithoutNewBlockReceivedByRound(round)
	return uint64(maxRoundsWithoutNewBlockReceived)
}

func (bp *baseProcessor) saveExecutedData(header data.HeaderHandler) error {
	if !header.IsHeaderV3() {
		return nil
	}

	executionResults := header.GetExecutionResultsHandlers()
	for _, execResult := range executionResults {
		err := bp.saveReceiptsForExecutionResult(execResult)
		if err != nil {
			return err
		}

		err = bp.saveMiniBlocksFromExecutionResults(execResult)
		if err != nil {
			return err
		}
		err = bp.saveIntermediateTxs(execResult.GetHeaderHash())
		if err != nil {
			return err
		}
	}
	return nil
}

func (bp *baseProcessor) saveMiniBlocksFromExecutionResults(baseExecutionResult data.BaseExecutionResultHandler) error {
	miniBlockHeaderHandlers, err := common.GetMiniBlocksHeaderHandlersFromExecResult(baseExecutionResult)
	if err != nil {
		return err
	}

	return bp.putMiniBlocksIntoStorage(miniBlockHeaderHandlers)
}

func (bp *baseProcessor) putMiniBlocksIntoStorage(miniBlockHeaderHandlers []data.MiniBlockHeaderHandler) error {
	if len(miniBlockHeaderHandlers) == 0 {
		return nil
	}

	miniBlockStorer, err := bp.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return err
	}

	executedMiniBlocksCache := bp.dataPool.ExecutedMiniBlocks()
	for _, miniBlockHeaderHandler := range miniBlockHeaderHandlers {
		mbHash := miniBlockHeaderHandler.GetHash()
		// do not save the cross-shard incoming mini blocks
		selfShardID := bp.shardCoordinator.SelfId()
		isCrossShardIncoming := miniBlockHeaderHandler.GetReceiverShardID() == selfShardID &&
			miniBlockHeaderHandler.GetSenderShardID() != selfShardID
		if isCrossShardIncoming {
			// no need to move into storer, should be there already
			executedMiniBlocksCache.Remove(mbHash)
			continue
		}

		cachedMiniBlock, found := executedMiniBlocksCache.Get(mbHash)
		if !found {
			log.Warn("mini block from execution result not cached after execution",
				"mini block hash", mbHash)
			return process.ErrMissingMiniBlock
		}

		cachedMiniBlockBytes := cachedMiniBlock.([]byte)
		errPut := miniBlockStorer.Put(mbHash, cachedMiniBlockBytes)
		if errPut != nil {
			return errPut
		}

		// mini block moved, cleaning the cache
		executedMiniBlocksCache.Remove(mbHash)
	}

	return nil
}

func (bp *baseProcessor) cacheIntraShardMiniBlocks(headerHash []byte, mbs block.MiniBlockSlice) error {
	marshalledMbs, err := bp.marshalizer.Marshal(&block.Body{MiniBlocks: mbs})
	if err != nil {
		return err
	}

	bp.dataPool.ExecutedMiniBlocks().Put(headerHash, marshalledMbs, len(marshalledMbs))

	return nil
}

func (bp *baseProcessor) saveReceiptsForHeader(header data.HeaderHandler, headerHash []byte) error {
	miniBlocks := bp.txCoordinator.GetCreatedInShardMiniBlocks()
	if len(miniBlocks) == 0 {
		return nil
	}

	receiptsHolder := holders.NewReceiptsHolder(miniBlocks)
	return bp.receiptsRepository.SaveReceipts(receiptsHolder, header, headerHash)
}

func (bp *baseProcessor) saveReceiptsForExecutionResult(
	execResult data.BaseExecutionResultHandler,
) error {
	miniBlocks, err := bp.getMiniBlocksForReceiptsV3(execResult)
	if err != nil {
		return err
	}

	if len(miniBlocks) == 0 {
		return nil
	}

	receiptsHolder := holders.NewReceiptsHolder(miniBlocks)
	return bp.receiptsRepository.SaveReceiptsForExecResult(receiptsHolder, execResult)
}

func (bp *baseProcessor) getMiniBlocksForReceiptsV3(execResult data.BaseExecutionResultHandler) ([]*block.MiniBlock, error) {
	headerHash := execResult.GetHeaderHash()

	executedMiniBlocksCache := bp.dataPool.ExecutedMiniBlocks()
	receiptsMiniBlocks, err := common.GetCachedMbs(executedMiniBlocksCache, bp.marshalizer, headerHash)
	if err != nil {
		return make([]*block.MiniBlock, 0), err
	}

	return receiptsMiniBlocks, nil
}

func (bp *baseProcessor) cacheLogEvents(headerHash []byte, logs []*data.LogData) error {
	logsMarshalled, err := bp.marshalizer.Marshal(logs)
	if err != nil {
		return err
	}

	key := common.PrepareLogEventsKey(headerHash)
	bp.dataPool.PostProcessTransactions().Put(key, logs, len(logsMarshalled))

	return nil
}

func (bp *baseProcessor) cacheExecutedMiniBlocks(body *block.Body, miniBlockHeaders []data.MiniBlockHeaderHandler) error {
	for i, mbHeader := range miniBlockHeaders {
		miniBlockHash := mbHeader.GetHash()
		marshalledMiniBlock, err := bp.marshalizer.Marshal(body.MiniBlocks[i])
		if err != nil {
			return err
		}

		bp.dataPool.ExecutedMiniBlocks().Put(miniBlockHash, marshalledMiniBlock, len(marshalledMiniBlock))
	}

	return nil
}

func (bp *baseProcessor) cacheIntermediateTxsForHeader(headerHash []byte) error {
	intermediateTxs := bp.txCoordinator.GetAllIntermediateTxs()
	buff, err := bp.marshalizer.Marshal(intermediateTxs)
	if err != nil {
		return err
	}

	bp.dataPool.PostProcessTransactions().Put(headerHash, intermediateTxs, len(buff))

	return nil
}

func (bp *baseProcessor) saveIntermediateTxs(headerHash []byte) error {
	postProcessTxsCache := bp.dataPool.PostProcessTransactions()
	cachedIntermediateTxs, ok := postProcessTxsCache.Get(headerHash)
	if !ok {
		log.Warn("saveIntermediateTxs: intermediateTxs not found in dataPool", "hash", headerHash)
		return fmt.Errorf("%w for header %s", process.ErrMissingHeader, hex.EncodeToString(headerHash))
	}

	cachedIntermediateTxsMap, ok := cachedIntermediateTxs.(map[block.Type]map[string]data.TransactionHandler)
	if !ok {
		log.Warn("saveIntermediateTxs: intermediateTxs cannot cast to concrete type", "hash", headerHash)
	}

	for blockType, cachedTransactionsMap := range cachedIntermediateTxsMap {
		err := bp.putTransactionsIntoStorage(blockType, cachedTransactionsMap)
		if err != nil {
			return err
		}
	}

	// all transactions moved, cleaning the cache
	postProcessTxsCache.Remove(headerHash)

	return nil
}

func (bp *baseProcessor) putTransactionsIntoStorage(blockType block.Type, cachedTransactionsMap map[string]data.TransactionHandler) error {
	unit, err := getStorageUnitFromBlockType(blockType)
	if err != nil {
		return err
	}

	storer, errGetStorer := bp.store.GetStorer(unit)
	if errGetStorer != nil {
		return errGetStorer
	}

	for txHash, txHandler := range cachedTransactionsMap {
		err = bp.putOneTransactionIntoStorage(storer, []byte(txHash), txHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bp *baseProcessor) putOneTransactionIntoStorage(
	storer storage.Storer,
	txHash []byte,
	tx data.TransactionHandler,
) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	buff, errMarshal := bp.marshalizer.Marshal(tx)
	if errMarshal != nil {
		return errMarshal
	}

	return storer.Put(txHash, buff)
}

func getStorageUnitFromBlockType(blockType block.Type) (dataRetriever.UnitType, error) {
	switch blockType {
	case block.TxBlock, block.InvalidBlock:
		return dataRetriever.TransactionUnit, nil
	case block.SmartContractResultBlock:
		return dataRetriever.UnsignedTransactionUnit, nil
	case block.ReceiptBlock:
		return dataRetriever.ReceiptsUnit, nil
	case block.RewardsBlock:
		return dataRetriever.UnsignedTransactionUnit, nil
	}
	return 0, process.ErrInvalidBlockType
}

func (bp *baseProcessor) checkInclusionEstimationForExecutionResults(header data.HeaderHandler) error {
	prevBlockLastExecutionResult, err := process.GetPrevBlockLastExecutionResult(bp.blockChain)
	if err != nil {
		return err
	}

	lastResultData, err := process.CreateDataForInclusionEstimation(prevBlockLastExecutionResult)
	if err != nil {
		return err
	}
	executionResults := header.GetExecutionResultsHandlers()
	allowed := bp.executionResultsInclusionEstimator.Decide(lastResultData, executionResults, header.GetRound())
	if allowed != len(executionResults) {
		log.Warn("number of execution results included in the header is not correct",
			"expected", allowed,
			"actual", len(executionResults),
		)
		return process.ErrInvalidNumberOfExecutionResultsInHeader
	}

	return nil
}

func (bp *baseProcessor) addExecutionResultsOnHeader(header data.HeaderHandler) error {
	pendingExecutionResults, err := bp.executionManager.GetPendingExecutionResults()
	if err != nil {
		return err
	}

	lastExecutionResultHandler, err := process.GetPrevBlockLastExecutionResult(bp.blockChain)
	if err != nil {
		return err
	}

	lastNotarizedExecutionResultInfo, err := process.CreateDataForInclusionEstimation(lastExecutionResultHandler)
	if err != nil {
		return err
	}

	var lastExecutionResultForCurrentBlock data.LastExecutionResultHandler
	numToInclude := bp.executionResultsInclusionEstimator.Decide(lastNotarizedExecutionResultInfo, pendingExecutionResults, header.GetRound())

	executionResultsToInclude := pendingExecutionResults[:numToInclude]
	lastExecutionResultForCurrentBlock = lastExecutionResultHandler
	if len(executionResultsToInclude) > 0 {
		lastExecutionResult := executionResultsToInclude[len(executionResultsToInclude)-1]
		lastExecutionResultForCurrentBlock, err = process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), lastExecutionResult, bp.shardCoordinator.SelfId())
		if err != nil {
			return err
		}
	}

	err = header.SetLastExecutionResultHandler(lastExecutionResultForCurrentBlock)
	if err != nil {
		return err
	}

	return header.SetExecutionResultsHandlers(executionResultsToInclude)
}

func (bp *baseProcessor) createMbsCrossShardDstMe(
	currentBlockHash []byte,
	currentBlock data.HeaderHandler,
	miniBlockProcessingInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (*CrossShardIncomingMbsCreationResult, error) {
	currMiniBlocksAdded, pendingMiniBlocks, currNumTxsAdded, hdrFinished, errCreate := bp.txCoordinator.CreateMbsCrossShardDstMe(
		currentBlock,
		miniBlockProcessingInfo,
	)
	if errCreate != nil {
		return nil, errCreate
	}

	if !hdrFinished {
		log.Debug("block cannot be fully processed",
			"round", currentBlock.GetRound(),
			"nonce", currentBlock.GetNonce(),
			"hash", currentBlockHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)
	}

	return &CrossShardIncomingMbsCreationResult{
		HeaderFinished:    hdrFinished,
		PendingMiniBlocks: pendingMiniBlocks,
		AddedMiniBlocks:   currMiniBlocksAdded,
	}, nil
}

func (bp *baseProcessor) revertGasForCrossShardDstMeMiniBlocks(added, pending []block.MiniblockAndHash) {
	miniBlockHashesToRevert := make([][]byte, 0, len(added))
	for _, mbAndHash := range added {
		miniBlockHashesToRevert = append(miniBlockHashesToRevert, mbAndHash.Hash)
	}
	for _, mbAndHash := range pending {
		miniBlockHashesToRevert = append(miniBlockHashesToRevert, mbAndHash.Hash)
	}

	bp.gasComputation.RevertIncomingMiniBlocks(miniBlockHashesToRevert)
}

func (bp *baseProcessor) setCurrentBlockInfo(
	header data.HeaderHandler,
	headerHash []byte,
	rootHash []byte,
) error {
	if header.IsHeaderV3() {
		// last executed info and header will be set on headers executor in async mode
		return bp.blockChain.SetCurrentBlockHeader(header)
	}

	err := bp.blockChain.SetCurrentBlockHeaderAndRootHash(header, rootHash)
	if err != nil {
		return err
	}

	// set also last executed block info and header
	// this will be useful at transition to Supernova with headers v3
	bp.blockChain.SetLastExecutedBlockHeaderAndRootHash(header, headerHash, rootHash)

	return nil
}

func (bp *baseProcessor) getLastExecutedRootHash(
	header data.HeaderHandler,
) []byte {
	rootHash := bp.getRootHash()
	if !header.IsHeaderV3() {
		return rootHash
	}

	lastExecutionResult, err := common.GetLastBaseExecutionResultHandler(header)
	if err != nil {
		log.Warn("failed to get last execution result for header", "err", err)
		return rootHash
	}

	return lastExecutionResult.GetRootHash()
}

// requestProofIfNeeded will request proof if Andromeda flag activated and not already in pool
func (bp *baseProcessor) requestProofIfNeeded(
	nonce uint64,
	shardID uint32,
	epoch uint32,
) {
	if !bp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return
	}

	proofsPool := bp.dataPool.Proofs()
	_, err := proofsPool.GetProofByNonce(nonce, shardID)
	if err == nil {
		// proof already in pool, no need to request it
		return
	}

	bp.requestHandler.RequestEquivalentProofByNonce(shardID, nonce)
}

func (bp *baseProcessor) requestHeadersFromHeaderIfNeeded(
	lastHeader data.HeaderHandler,
) {
	lastRound := lastHeader.GetRound()
	shardID := lastHeader.GetShardID()

	shouldRequestCrossHeaders := bp.roundHandler.Index() > int64(lastRound+bp.getMaxRoundsWithoutBlockReceived(lastRound))

	if !shouldRequestCrossHeaders {
		return
	}

	fromNonce := lastHeader.GetNonce() + 1
	toNonce := fromNonce + numHeadersToRequestInAdvance

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		bp.requestHeaderIfNeeded(nonce, shardID)
		bp.requestProofIfNeeded(nonce, shardID, lastHeader.GetEpoch())
	}
}

func (bp *baseProcessor) requestHeaderIfNeeded(
	nonce uint64,
	shardID uint32,
) {
	headersPool := bp.dataPool.Headers()
	_, _, err := headersPool.GetHeadersByNonceAndShardId(nonce, shardID)
	if err == nil {
		// header already in pool, no need to request it
		return
	}

	bp.requestHeaderByShardAndNonce(shardID, nonce)
}

func (bp *baseProcessor) verifyGasLimit(header data.HeaderHandler) error {
	splitRes, err := bp.splitTransactionsForHeader(header)
	if err != nil {
		return err
	}

	err = bp.checkMetaOutgoingResults(header, splitRes)
	if err != nil {
		return err
	}

	bp.gasComputation.Reset()
	_, numPendingMiniBlocks, err := bp.gasComputation.AddIncomingMiniBlocks(splitRes.incomingMiniBlocks, splitRes.incomingTransactions)
	if err != nil {
		return err
	}

	// for meta, both splitRes.outgoingTransactionHashes and splitRes.outgoingTransactions should be empty, checked on checkMetaOutgoingResults
	addedTxHashes, pendingMiniBlocksAdded, err := bp.gasComputation.AddOutgoingTransactions(splitRes.outgoingTransactionHashes, splitRes.outgoingTransactions)
	if err != nil {
		return err
	}
	if len(addedTxHashes) != len(splitRes.outgoingTransactionHashes) {
		return fmt.Errorf("%w, outgoing transactions exceeded the limit", process.ErrInvalidMaxGasLimitPerMiniBlock)
	}

	if numPendingMiniBlocks != len(pendingMiniBlocksAdded) {
		return fmt.Errorf("%w, incoming mini blocks exceeded the limit", process.ErrInvalidMaxGasLimitPerMiniBlock)
	}

	return nil
}

func (bp *baseProcessor) checkMetaOutgoingResults(
	header data.HeaderHandler,
	splitRes *splitTxsResult,
) error {
	_, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return nil
	}

	numOutGoingMBs := len(splitRes.outGoingMiniBlocks)
	if numOutGoingMBs != 0 {
		return fmt.Errorf("%w, received: %d", errInvalidNumOutGoingMBInMetaHdrProposal, numOutGoingMBs)
	}

	numOutGoingTxs := len(splitRes.outgoingTransactions)
	if numOutGoingTxs != 0 {
		return fmt.Errorf("%w in metaProcessor.verifyGasLimit, received: %d",
			errInvalidNumOutGoingTxsInMetaHdrProposal,
			numOutGoingTxs,
		)
	}

	return nil
}

func (bp *baseProcessor) splitTransactionsForHeader(header data.HeaderHandler) (*splitTxsResult, error) {
	incomingMiniBlocks := make([]data.MiniBlockHeaderHandler, 0)
	outGoingMiniBlocks := make([]data.MiniBlockHeaderHandler, 0)
	outgoingTransactionHashes := make([][]byte, 0)
	incomingTransactions := make(map[string][]data.TransactionHandler)
	outgoingTransactions := make([]data.TransactionHandler, 0)
	for _, mb := range header.GetMiniBlockHeaderHandlers() {
		txHashes, txsForMb, err := bp.getTransactionsForMiniBlock(mb)
		if err != nil {
			return nil, err
		}

		if mb.GetSenderShardID() == bp.shardCoordinator.SelfId() {
			outgoingTransactionHashes = append(outgoingTransactionHashes, txHashes...)
			outgoingTransactions = append(outgoingTransactions, txsForMb...)
			outGoingMiniBlocks = append(outGoingMiniBlocks, mb)
			continue
		}

		incomingMiniBlocks = append(incomingMiniBlocks, mb)
		incomingTransactions[string(mb.GetHash())] = txsForMb
	}

	return &splitTxsResult{
		incomingMiniBlocks:        incomingMiniBlocks,
		outGoingMiniBlocks:        outGoingMiniBlocks,
		incomingTransactions:      incomingTransactions,
		outgoingTransactionHashes: outgoingTransactionHashes,
		outgoingTransactions:      outgoingTransactions,
	}, nil
}

func (bp *baseProcessor) getTransactionsForMiniBlock(
	miniBlock data.MiniBlockHeaderHandler,
) ([][]byte, []data.TransactionHandler, error) {
	obj, hashInPool := bp.dataPool.MiniBlocks().Get(miniBlock.GetHash())
	if !hashInPool {
		return nil, nil, process.ErrMissingMiniBlock
	}

	mbForHeaderPtr, typeOk := obj.(*block.MiniBlock)
	if !typeOk {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	txs := make([]data.TransactionHandler, len(mbForHeaderPtr.TxHashes))
	var err error
	for idx, txHash := range mbForHeaderPtr.TxHashes {
		txs[idx], err = process.GetTransactionHandlerFromPool(
			miniBlock.GetSenderShardID(),
			miniBlock.GetReceiverShardID(),
			txHash,
			bp.dataPool.Transactions(),
			process.SearchMethodSearchFirst,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	return mbForHeaderPtr.TxHashes, txs, nil
}

// getCurrentBlockHeader returns the current block header from blockchain.
func (bp *baseProcessor) getCurrentBlockHeader() data.HeaderHandler {
	currentBlockHeader := bp.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(currentBlockHeader) {
		return currentBlockHeader
	}

	return bp.blockChain.GetGenesisHeader()
}

func (bp *baseProcessor) getLastExecutionResultHeader(
	currentHeader data.HeaderHandler,
) (data.HeaderHandler, error) {
	if !currentHeader.IsHeaderV3() {
		return currentHeader, nil
	}

	lastExecutionResult, err := common.GetLastBaseExecutionResultHandler(currentHeader)
	if err != nil {
		return nil, err
	}

	headersPool := bp.dataPool.Headers()

	header, err := headersPool.GetHeaderByHash(lastExecutionResult.GetHeaderHash())
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (bp *baseProcessor) checkContextBeforeExecution(header data.HeaderHandler) error {
	lastCommittedRootHash, err := bp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	lastExecutedNonce, lastExecutedHash, lastExecutedRootHash := bp.blockChain.GetLastExecutedBlockInfo()
	if !bytes.Equal(header.GetPrevHash(), lastExecutedHash) {
		return process.ErrBlockHashDoesNotMatch
	}
	if header.GetNonce() != lastExecutedNonce+1 {
		return process.ErrWrongNonceInBlock
	}
	if !bytes.Equal(lastCommittedRootHash, lastExecutedRootHash) {
		return process.ErrRootStateDoesNotMatch
	}

	return nil
}

func (bp *baseProcessor) getCrossShardIncomingMiniBlocksFromBody(body *block.Body) []*block.MiniBlock {
	miniBlocks := make([]*block.MiniBlock, 0)

	isReceiverCurrentShard := false
	for _, mb := range body.MiniBlocks {
		isReceiverCurrentShard = mb.ReceiverShardID == bp.shardCoordinator.SelfId() || mb.ReceiverShardID == core.AllShardId
		if isReceiverCurrentShard && mb.SenderShardID != bp.shardCoordinator.SelfId() {
			miniBlocks = append(miniBlocks, mb)
		}
	}
	return miniBlocks
}

func (bp *baseProcessor) collectMiniBlocks(
	headerHash []byte,
	body *block.Body,
) ([]data.MiniBlockHeaderHandler, int, []byte, error) {
	crossShardIncomingMiniBlocks := bp.getCrossShardIncomingMiniBlocksFromBody(body)
	miniBlocksFromSelf := bp.txCoordinator.GetCreatedMiniBlocksFromMe()
	postProcessMiniBlocks := bp.txCoordinator.CreatePostProcessMiniBlocks()

	allMiniBlocks := make([]*block.MiniBlock, 0, len(crossShardIncomingMiniBlocks)+len(miniBlocksFromSelf)+len(postProcessMiniBlocks))
	allMiniBlocks = append(allMiniBlocks, crossShardIncomingMiniBlocks...)
	allMiniBlocks = append(allMiniBlocks, miniBlocksFromSelf...)
	allMiniBlocks = append(allMiniBlocks, postProcessMiniBlocks...)

	receiptHash, err := bp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, 0, nil, err
	}

	bodyAfterExecution := &block.Body{MiniBlocks: allMiniBlocks}
	// remove the self-receipts and self smart contract results mini blocks - similar to Pre-Supernova
	sanitizedBodyAfterExecution := deleteSelfReceiptsMiniBlocks(bodyAfterExecution)

	// giving an empty processedMiniBlockInfo would cause all miniBlockHeaders to be created as fully processed.
	processedMiniBlockInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	totalTxCount, miniBlockHeaderHandlers, err := bp.createMiniBlockHeaderHandlers(sanitizedBodyAfterExecution, processedMiniBlockInfo)
	if err != nil {
		return nil, 0, nil, err
	}

	intraMiniBlocks := bp.txCoordinator.GetCreatedInShardMiniBlocks()
	err = bp.cacheIntraShardMiniBlocks(headerHash, intraMiniBlocks)
	if err != nil {
		return nil, 0, nil, err
	}

	err = bp.cacheExecutedMiniBlocks(sanitizedBodyAfterExecution, miniBlockHeaderHandlers)
	if err != nil {
		return nil, 0, nil, err
	}

	return miniBlockHeaderHandlers, totalTxCount, receiptHash, nil
}

func (bp *baseProcessor) getBlockBodyFromPool(
	header data.HeaderHandler,
	miniBlockHeaderHandlers []data.MiniBlockHeaderHandler,
) (data.BodyHandler, error) {
	miniBlocksPool := bp.dataPool.MiniBlocks()
	var miniBlocks block.MiniBlockSlice

	for _, mbHeader := range miniBlockHeaderHandlers {
		obj, hashInPool := miniBlocksPool.Get(mbHeader.GetHash())
		if !hashInPool {
			continue
		}

		miniBlock, typeOk := obj.(*block.MiniBlock)
		if !typeOk {
			return nil, process.ErrWrongTypeAssertion
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func getProposedAndExecutedMiniBlockHeaders(
	header data.HeaderHandler,
) ([]data.MiniBlockHeaderHandler, error) {
	if !header.IsHeaderV3() {
		return header.GetMiniBlockHeaderHandlers(), nil
	}

	execResultsMiniBlockHeaders, err := common.GetMiniBlockHeadersFromExecResult(header)
	if err != nil {
		return nil, err
	}

	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	miniBlockHeaders = append(miniBlockHeaders, execResultsMiniBlockHeaders...)

	return miniBlockHeaders, nil
}
