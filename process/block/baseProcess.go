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
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/block")

type hashAndHdr struct {
	hdr  data.HeaderHandler
	hash []byte
}

type nonceAndHashInfo struct {
	hash  []byte
	nonce uint64
}

type hdrInfo struct {
	usedInBlock bool
	hdr         data.HeaderHandler
}

type baseProcessor struct {
	shardCoordinator        sharding.Coordinator
	nodesCoordinator        nodesCoordinator.NodesCoordinator
	accountsDB              map[state.AccountsDbIdentifier]state.AccountsAdapter
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
	hdrsForCurrBlock        *hdrForBlock
	genesisNonce            uint64
	mutProcessDebugger      sync.RWMutex
	processDebugger         process.Debugger
	processStatusHandler    common.ProcessStatusHandler
	managedPeersHolder      common.ManagedPeersHolder

	versionedHeaderFactory       nodeFactory.VersionedHeaderFactory
	headerIntegrityVerifier      process.HeaderIntegrityVerifier
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	blockProcessingCutoffHandler cutoff.BlockProcessingCutoffHandler

	appStatusHandler core.AppStatusHandler
	blockProcessor   blockProcessor
	txCounter        *transactionCounter

	outportHandler      outport.OutportHandler
	outportDataProvider outport.DataProviderOutport
	historyRepo         dblookupext.HistoryRepository
	epochNotifier       process.EpochNotifier
	enableEpochsHandler common.EnableEpochsHandler
	roundNotifier       process.RoundNotifier
	enableRoundsHandler process.EnableRoundsHandler
	vmContainerFactory  process.VirtualMachinesContainerFactory
	vmContainer         process.VirtualMachinesContainer
	gasConsumedProvider gasConsumedProvider
	economicsData       process.EconomicsDataHandler

	processDataTriesOnCommitEpoch bool
	lastRestartNonce              uint64
	pruningDelay                  uint32
	processedMiniBlocksTracker    process.ProcessedMiniBlocksTracker
	receiptsRepository            receiptsRepository

	mutNonceOfFirstCommittedBlock sync.RWMutex
	nonceOfFirstCommittedBlock    core.OptionalUint64
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
		go bp.requestHeaderByShardAndNonce(shardId, nonce)
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

func displayHeader(headerHandler data.HeaderHandler) []*display.LineData {
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
	prevAggregatedSig, prevBitmap := headerHandler.GetPreviousAggregatedSignatureAndBitmap()
	return []*display.LineData{
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
		display.NewLineData(false, []string{
			"",
			"Epoch start meta hash",
			logger.DisplayByteSlice(epochStartMetaHash)}),
		display.NewLineData(false, []string{
			"Previous proof",
			"Aggregated signature",
			logger.DisplayByteSlice(prevAggregatedSig)}),
		display.NewLineData(true, []string{
			"",
			"Pub keys bitmap",
			logger.DisplayByteSlice(prevBitmap)}),
	}
}

// checkProcessorParameters will check the input parameters values
func checkProcessorParameters(arguments ArgBaseProcessor) error {

	for key := range arguments.AccountsDB {
		if check.IfNil(arguments.AccountsDB[key]) {
			return process.ErrNilAccountsAdapter
		}
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
	})
	if err != nil {
		return err
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

	return nil
}

func (bp *baseProcessor) createBlockStarted() error {
	bp.hdrsForCurrBlock.resetMissingHdrs()
	bp.hdrsForCurrBlock.initMaps()
	scheduledGasAndFees := bp.scheduledTxsExecutionHandler.GetScheduledGasAndFees()
	bp.txCoordinator.CreateBlockStarted()
	bp.feeHandler.CreateBlockStarted(scheduledGasAndFees)

	err := bp.txCoordinator.AddIntermediateTransactions(bp.scheduledTxsExecutionHandler.GetScheduledIntermediateTxs())
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
func (bp *baseProcessor) sortHeadersForCurrentBlockByNonce(usedInBlock bool) map[uint32][]data.HeaderHandler {
	hdrsForCurrentBlock := make(map[uint32][]data.HeaderHandler)

	bp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, headerInfo := range bp.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlock[headerInfo.hdr.GetShardID()] = append(hdrsForCurrentBlock[headerInfo.hdr.GetShardID()], headerInfo.hdr)
	}
	bp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	// sort headers for each shard
	for _, hdrsForShard := range hdrsForCurrentBlock {
		process.SortHeadersByNonce(hdrsForShard)
	}

	return hdrsForCurrentBlock
}

func (bp *baseProcessor) sortHeaderHashesForCurrentBlockByNonce(usedInBlock bool) map[uint32][][]byte {
	hdrsForCurrentBlockInfo := make(map[uint32][]*nonceAndHashInfo)

	bp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for metaBlockHash, headerInfo := range bp.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlockInfo[headerInfo.hdr.GetShardID()] = append(hdrsForCurrentBlockInfo[headerInfo.hdr.GetShardID()],
			&nonceAndHashInfo{nonce: headerInfo.hdr.GetNonce(), hash: []byte(metaBlockHash)})
	}
	bp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for _, hdrsForShard := range hdrsForCurrentBlockInfo {
		if len(hdrsForShard) > 1 {
			sort.Slice(hdrsForShard, func(i, j int) bool {
				return hdrsForShard[i].nonce < hdrsForShard[j].nonce
			})
		}
	}

	hdrsHashesForCurrentBlock := make(map[uint32][][]byte, len(hdrsForCurrentBlockInfo))
	for shardId, hdrsForShard := range hdrsForCurrentBlockInfo {
		for _, hdrForShard := range hdrsForShard {
			hdrsHashesForCurrentBlock[shardId] = append(hdrsHashesForCurrentBlock[shardId], hdrForShard.hash)
		}
	}

	return hdrsHashesForCurrentBlock
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

// check if header has the same miniblocks as presented in body
func (bp *baseProcessor) checkHeaderBodyCorrelation(miniBlockHeaders []data.MiniBlockHeaderHandler, body *block.Body) error {
	mbHashesFromHdr := make(map[string]data.MiniBlockHeaderHandler, len(miniBlockHeaders))
	for i := 0; i < len(miniBlockHeaders); i++ {
		mbHashesFromHdr[string(miniBlockHeaders[i].GetHash())] = miniBlockHeaders[i]
	}

	if len(miniBlockHeaders) != len(body.MiniBlocks) {
		return process.ErrHeaderBodyMismatch
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock == nil {
			return process.ErrNilMiniBlock
		}

		mbHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, miniBlock)
		if err != nil {
			return err
		}

		mbHdr, ok := mbHashesFromHdr[string(mbHash)]
		if !ok {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.GetTxCount() != uint32(len(miniBlock.TxHashes)) {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.GetReceiverShardID() != miniBlock.ReceiverShardID {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.GetSenderShardID() != miniBlock.SenderShardID {
			return process.ErrHeaderBodyMismatch
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

// requestMissingFinalityAttestingHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the finality headers greater than the highest header, for given shard,
// related to the block which should be processed
// this method should be called only under the mutex protection: hdrsForCurrBlock.mutHdrsForBlock
func (bp *baseProcessor) requestMissingFinalityAttestingHeaders(
	shardID uint32,
	finality uint32,
) uint32 {
	requestedHeaders := uint32(0)
	missingFinalityAttestingHeaders := uint32(0)

	highestHdrNonce := bp.hdrsForCurrBlock.highestHdrNonce[shardID]
	if highestHdrNonce == uint64(0) {
		return missingFinalityAttestingHeaders
	}

	headersPool := bp.dataPool.Headers()
	lastFinalityAttestingHeader := highestHdrNonce + uint64(finality)
	for i := highestHdrNonce + 1; i <= lastFinalityAttestingHeader; i++ {
		headers, headersHashes, err := headersPool.GetHeadersByNonceAndShardId(i, shardID)
		if err != nil {
			missingFinalityAttestingHeaders++
			requestedHeaders++
			go bp.requestHeaderByShardAndNonce(shardID, i)
			continue
		}

		for index := range headers {
			bp.hdrsForCurrBlock.hdrHashAndInfo[string(headersHashes[index])] = &hdrInfo{
				hdr:         headers[index],
				usedInBlock: false,
			}
		}
	}

	if requestedHeaders > 0 {
		log.Debug("requested missing finality attesting headers",
			"num headers", requestedHeaders,
			"shard", shardID)
	}

	return missingFinalityAttestingHeaders
}

func (bp *baseProcessor) requestHeaderByShardAndNonce(shardID uint32, nonce uint64) {
	if shardID == core.MetachainShardId {
		bp.requestHandler.RequestMetaHeaderByNonce(nonce)
	} else {
		bp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
	}
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
		highestPrevFinalBlockNonce)

	if bp.shardCoordinator.SelfId() == core.MetachainShardId {
		for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
			bp.cleanupPoolsForCrossShard(shardID, noncesToPrevFinal)
		}
	} else {
		bp.cleanupPoolsForCrossShard(core.MetachainShardId, noncesToPrevFinal)
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

	bp.removeHeadersBehindNonceFromPools(
		false,
		shardID,
		crossNotarizedHeader.GetNonce(),
	)
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
	bodyHandler, err := bp.requestBlockBodyHandler.GetBlockBodyFromPool(headerHandler)
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

	return bp.txCoordinator.RemoveTxsFromPool(newBody)
}

func (bp *baseProcessor) getFinalMiniBlocks(header data.HeaderHandler, body *block.Body) (*block.Body, error) {
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

	selfNotarizedNonce := selfNotarizedHeader.GetNonce()

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

		crossNotarizedNonce = crossNotarizedHeader.GetNonce()
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
	finalBlockNonce := bp.forkDetector.GetHighestFinalBlockNonce()
	if currentBlockNonce > finalBlockNonce {
		noncesToFinal = currentBlockNonce - finalBlockNonce
	}

	return noncesToFinal
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

	bp.txCoordinator.SaveTxsToStorage(body)
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

	receiptsHolder := holders.NewReceiptsHolder(bp.txCoordinator.GetCreatedInShardMiniBlocks())
	errNotCritical = bp.receiptsRepository.SaveReceipts(receiptsHolder, header, headerHash)
	if errNotCritical != nil {
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

func (bp *baseProcessor) saveShardHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	startTime := time.Now()

	nonceToByteSlice := bp.uint64Converter.ToByteSlice(header.GetNonce())
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(header.GetShardID())

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

	elapsedTime := time.Since(startTime)
	if elapsedTime >= common.PutInStorerMaxTime {
		log.Warn("saveMetaHeader", "elapsed time", elapsedTime)
	}
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
	finalHeader data.HeaderHandler,
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
	accounts.PruneTrie(prevRootHash, state.OldRoot, bp.getPruningHandler(finalHeader.GetNonce()))
}

// RevertCurrentBlock reverts the current block for cleanup failed process
func (bp *baseProcessor) RevertCurrentBlock() {
	bp.revertAccountState()
	bp.revertScheduledInfo()
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

func (bp *baseProcessor) commitAll(headerHandler data.HeaderHandler) error {
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

func (bp *baseProcessor) requestMiniBlocksIfNeeded(headerHandler data.HeaderHandler) {
	lastCrossNotarizedHeader, _, err := bp.blockTracker.GetLastCrossNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Debug("requestMiniBlocksIfNeeded.GetLastCrossNotarizedHeader",
			"shard", headerHandler.GetShardID(),
			"error", err.Error())
		return
	}

	isHeaderOutOfRequestRange := headerHandler.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToRequestInAdvance
	if isHeaderOutOfRequestRange {
		return
	}

	waitTime := common.ExtraDelayForRequestBlockInfo
	roundDifferences := bp.roundHandler.Index() - int64(headerHandler.GetRound())
	if roundDifferences > 1 {
		waitTime = 0
	}

	// waiting for late broadcast of mini blocks and transactions to be done and received
	time.Sleep(waitTime)

	bp.txCoordinator.RequestMiniBlocksAndTransactions(headerHandler)
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

func (bp *baseProcessor) commitTrieEpochRootHashIfNeeded(metaBlock *block.MetaBlock, rootHash []byte) error {
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

	epochBytes := bp.uint64Converter.ToByteSlice(uint64(metaBlock.Epoch))

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
		Hasher:              bp.hasher,
		Marshaller:          bp.marshalizer,
		EnableEpochsHandler: bp.enableEpochsHandler,
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
		"epoch", metaBlock.Epoch,
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
			bp.RevertCurrentBlock()
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
