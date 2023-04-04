package block

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var rootHash = "uncomputed root hash"

type extendedShardHeaderTrackHandler interface {
	ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
}

type sovereignBlockProcessor struct {
	*shardProcessor
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	uncomputedRootHash           []byte
	extendedShardHeaderTracker   extendedShardHeaderTrackHandler
}

// NewSovereignBlockProcessor creates a new sovereign block processor
func NewSovereignBlockProcessor(
	shardProcessor *shardProcessor,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
) (*sovereignBlockProcessor, error) {

	sbp := &sovereignBlockProcessor{
		shardProcessor:               shardProcessor,
		validatorStatisticsProcessor: validatorStatisticsProcessor,
	}

	sbp.uncomputedRootHash = sbp.hasher.Compute(rootHash)

	extendedShardHeaderTracker, ok := sbp.blockTracker.(extendedShardHeaderTrackHandler)
	if !ok {
		return nil, fmt.Errorf("%w in NewSovereignBlockProcessor", process.ErrWrongTypeAssertion)
	}

	sbp.extendedShardHeaderTracker = extendedShardHeaderTracker

	return sbp, nil
}

// CreateNewHeader creates a new header
func (s *sovereignBlockProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	s.enableRoundsHandler.CheckRound(round)
	header := &block.SovereignChainHeader{
		Header: &block.Header{
			SoftwareVersion: process.SovereignHeaderVersion,
			RootHash:        s.uncomputedRootHash,
		},
		ValidatorStatsRootHash: s.uncomputedRootHash,
	}

	err := s.setRoundNonceInitFees(round, nonce, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// CreateBlock selects and puts transaction into the temporary block body
func (s *sovereignBlockProcessor) CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}

	sovereignChainHeaderHandler, ok := initialHdr.(data.SovereignChainHeaderHandler)
	if !ok {
		return nil, nil, fmt.Errorf("%w in sovereignBlockProcessor.CreateBlock", process.ErrWrongTypeAssertion)
	}

	s.processStatusHandler.SetBusy("sovereignBlockProcessor.CreateBlock")
	defer s.processStatusHandler.SetIdle()

	for _, accounts := range s.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignBlockProcessor.CreateBlock first entry", "stack", accounts.GetStackDebugFirstEntry())
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	err := s.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	s.blockChainHook.SetCurrentHeader(initialHdr)

	var miniBlocks block.MiniBlockSlice
	//processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	if !haveTime() {
		log.Debug("sovereignBlockProcessor.CreateBlock", "error", process.ErrTimeIsOut)

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return nil, nil, process.ErrTimeIsOut
	}

	startTime := time.Now()
	createIncomingMiniBlocksDestMeInfo, err := s.createIncomingMiniBlocksDestMe(haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)
	if err != nil {
		log.Debug("createIncomingMiniBlocksDestMe", "error", err.Error())
	}
	if createIncomingMiniBlocksDestMeInfo != nil {
		//processedMiniBlocksDestMeInfo = createIncomingMiniBlocksDestMeInfo.allProcessedMiniBlocksInfo
		if len(createIncomingMiniBlocksDestMeInfo.miniBlocks) > 0 {
			miniBlocks = append(miniBlocks, createIncomingMiniBlocksDestMeInfo.miniBlocks...)

			log.Debug("created mini blocks and txs with destination in self shard",
				"num mini blocks", len(createIncomingMiniBlocksDestMeInfo.miniBlocks),
				"num txs", createIncomingMiniBlocksDestMeInfo.numTxsAdded,
				"num extended shard headers", createIncomingMiniBlocksDestMeInfo.numHdrsAdded)
		}
	}

	startTime = time.Now()
	mbsFromMe := s.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, initialHdr.GetPrevRandSeed())
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to create mbs from me", "time", elapsedTime)

	if len(mbsFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbsFromMe...)

		numTxs := 0
		for _, mb := range mbsFromMe {
			numTxs += len(mb.TxHashes)
		}

		log.Debug("processed miniblocks and txs from self shard",
			"num miniblocks", len(mbsFromMe),
			"num txs", numTxs)
	}

	mainChainShardHeaderHashes := s.sortExtendedShardHeaderHashesForCurrentBlockByNonce(true)
	err = sovereignChainHeaderHandler.SetMainChainShardHeaderHashes(mainChainShardHeaderHashes)
	if err != nil {
		return nil, nil, err
	}

	return initialHdr, &block.Body{MiniBlocks: miniBlocks}, nil
}

func (s *sovereignBlockProcessor) createIncomingMiniBlocksDestMe(haveTime func() bool) (*createAndProcessMiniBlocksDestMeInfo, error) {
	log.Debug("createIncomingMiniBlocksDestMe has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestExtendedShardChainFromLastNotarized")
	orderedExtendedShardHeaders, orderedExtendedShardHeadersHashes, err := s.extendedShardHeaderTracker.ComputeLongestExtendedShardChainFromLastNotarized()
	sw.Stop("ComputeLongestExtendedShardChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return nil, err
	}

	log.Debug("extended shard headers ordered",
		"num extended shard headers", len(orderedExtendedShardHeaders),
	)

	lastExtendedShardHdr, _, err := s.blockTracker.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	haveAdditionalTimeFalse := func() bool {
		return false
	}

	createAndProcessInfo := &createAndProcessMiniBlocksDestMeInfo{
		haveTime:                   haveTime,
		haveAdditionalTime:         haveAdditionalTimeFalse,
		miniBlocks:                 make(block.MiniBlockSlice, 0),
		allProcessedMiniBlocksInfo: make(map[string]*processedMb.ProcessedMiniBlockInfo),
		numTxsAdded:                uint32(0),
		numHdrsAdded:               uint32(0),
		scheduledMode:              true,
	}

	// do processing in order
	s.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedExtendedShardHeadersHashes); i++ {
		if !createAndProcessInfo.haveTime() && !createAndProcessInfo.haveAdditionalTime() {
			log.Debug("time is up in creating incoming mini blocks destination me",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"num txs added", createAndProcessInfo.numTxsAdded,
			)
			break
		}

		if createAndProcessInfo.numHdrsAdded >= process.MaxExtendedShardHeadersAllowedInOneSovereignBlock {
			log.Debug("maximum extended shard headers allowed to be included in one sovereign block has been reached",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"extended shard headers added", createAndProcessInfo.numHdrsAdded,
			)
			break
		}

		extendedShardHeader, ok := orderedExtendedShardHeaders[i].(data.ShardHeaderExtendedHandler)
		if !ok {
			log.Debug("wrong type assertion from data.HeaderHandler to data.ShardHeaderExtendedHandler",
				"hash", orderedExtendedShardHeadersHashes[i],
				"shard", orderedExtendedShardHeaders[i].GetShardID(),
				"round", orderedExtendedShardHeaders[i].GetRound(),
				"nonce", orderedExtendedShardHeaders[i].GetNonce())
			break
		}

		createAndProcessInfo.currHdr = orderedExtendedShardHeaders[i]
		if createAndProcessInfo.currHdr.GetNonce() > lastExtendedShardHdr.GetNonce()+1 {
			log.Debug("skip searching",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"last extended shard hdr nonce", lastExtendedShardHdr.GetNonce(),
				"curr extended shard hdr nonce", createAndProcessInfo.currHdr.GetNonce())
			break
		}

		createAndProcessInfo.currHdrHash = orderedExtendedShardHeadersHashes[i]
		if len(extendedShardHeader.GetIncomingMiniBlockHandlers()) == 0 {
			s.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currHdrHash)] = &hdrInfo{hdr: createAndProcessInfo.currHdr, usedInBlock: true}
			createAndProcessInfo.numHdrsAdded++
			lastExtendedShardHdr = createAndProcessInfo.currHdr
			continue
		}

		createAndProcessInfo.currProcessedMiniBlocksInfo = s.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(createAndProcessInfo.currHdrHash)
		createAndProcessInfo.hdrAdded = false

		shouldContinue, errCreated := s.createIncomingMiniBlocksAndTransactionsDestMe(createAndProcessInfo)
		if errCreated != nil {
			return nil, errCreated
		}
		if !shouldContinue {
			break
		}

		lastExtendedShardHdr = createAndProcessInfo.currHdr
	}
	s.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	go s.requestExtendedShardHeadersIfNeeded(createAndProcessInfo.numHdrsAdded, lastExtendedShardHdr)

	for _, miniBlock := range createAndProcessInfo.miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("createIncomingMiniBlocksDestMe has been finished",
		"num txs added", createAndProcessInfo.numTxsAdded,
		"num hdrs added", createAndProcessInfo.numHdrsAdded)

	return createAndProcessInfo, nil
}

//TODO: This mock should be removed when real functionality will be implemented
var txCoordinatorMock = testscommon.TransactionCoordinatorMock{
	CreateMbsAndProcessCrossShardTransactionsDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo, haveTime func() bool, haveAdditionalTime func() bool, scheduledMode bool) (block.MiniBlockSlice, uint32, bool, error) {
		return make(block.MiniBlockSlice, 0), 0, false, nil
	},
}

func (s *sovereignBlockProcessor) createIncomingMiniBlocksAndTransactionsDestMe(
	createAndProcessInfo *createAndProcessMiniBlocksDestMeInfo,
) (bool, error) {
	//TODO: Replace this mock object with transaction coordinator object, when real functionality will be implemented
	currMiniBlocksAdded, currNumTxsAdded, hdrProcessFinished, errCreated := txCoordinatorMock.CreateMbsAndProcessCrossShardTransactionsDstMe(
		createAndProcessInfo.currHdr,
		createAndProcessInfo.currProcessedMiniBlocksInfo,
		createAndProcessInfo.haveTime,
		createAndProcessInfo.haveAdditionalTime,
		createAndProcessInfo.scheduledMode)
	if errCreated != nil {
		return false, errCreated
	}

	for miniBlockHash, processedMiniBlockInfo := range createAndProcessInfo.currProcessedMiniBlocksInfo {
		createAndProcessInfo.allProcessedMiniBlocksInfo[miniBlockHash] = &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         processedMiniBlockInfo.FullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}

	// all txs processed, add to processed miniblocks
	createAndProcessInfo.miniBlocks = append(createAndProcessInfo.miniBlocks, currMiniBlocksAdded...)
	createAndProcessInfo.numTxsAdded += currNumTxsAdded

	if !createAndProcessInfo.hdrAdded && currNumTxsAdded > 0 {
		s.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currHdrHash)] = &hdrInfo{hdr: createAndProcessInfo.currHdr, usedInBlock: true}
		createAndProcessInfo.numHdrsAdded++
		createAndProcessInfo.hdrAdded = true
	}

	if !hdrProcessFinished {
		log.Debug("extended shard header cannot be fully processed",
			"scheduled mode", createAndProcessInfo.scheduledMode,
			"round", createAndProcessInfo.currHdr.GetRound(),
			"nonce", createAndProcessInfo.currHdr.GetNonce(),
			"hash", createAndProcessInfo.currHdrHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)

		return false, nil
	}

	return true, nil
}

func (s *sovereignBlockProcessor) requestExtendedShardHeadersIfNeeded(hdrsAdded uint32, lastExtendedShardHdr data.HeaderHandler) {
	log.Debug("extended shard headers added",
		"num", hdrsAdded,
		"highest nonce", lastExtendedShardHdr.GetNonce(),
	)
	//TODO: A request mechanism should be implemented if extended shard header(s) is(are) needed
}

func (s *sovereignBlockProcessor) sortExtendedShardHeaderHashesForCurrentBlockByNonce(usedInBlock bool) [][]byte {
	hdrsForCurrentBlockInfo := make([]*nonceAndHashInfo, 0)

	s.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for headerHash, headerInfo := range s.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlockInfo = append(hdrsForCurrentBlockInfo,
			&nonceAndHashInfo{nonce: headerInfo.hdr.GetNonce(), hash: []byte(headerHash)})
	}
	s.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	if len(hdrsForCurrentBlockInfo) > 1 {
		sort.Slice(hdrsForCurrentBlockInfo, func(i, j int) bool {
			return hdrsForCurrentBlockInfo[i].nonce < hdrsForCurrentBlockInfo[j].nonce
		})
	}

	hdrsHashesForCurrentBlock := make([][]byte, len(hdrsForCurrentBlockInfo))
	for index, hdrForCurrentBlockInfo := range hdrsForCurrentBlockInfo {
		hdrsHashesForCurrentBlock[index] = hdrForCurrentBlockInfo.hash
	}

	return hdrsHashesForCurrentBlock
}

// ProcessBlock actually processes the selected transaction and will create the final block body
func (s *sovereignBlockProcessor) ProcessBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
	if haveTime == nil {
		return nil, nil, process.ErrNilHaveTimeHandler
	}

	s.processStatusHandler.SetBusy("sovereignBlockProcessor.ProcessBlock")
	defer s.processStatusHandler.SetIdle()

	err := s.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			go s.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return nil, nil, err
	}

	blockBody, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	err = s.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	s.txCoordinator.RequestBlockTransactions(blockBody)
	err = s.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return nil, nil, err
	}

	for _, accounts := range s.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignBlockProcessor.ProcessBlock first entry", "stack", accounts.GetStackDebugFirstEntry())
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	defer func() {
		if err != nil {
			s.RevertCurrentBlock()
		}
	}()

	startTime := time.Now()
	miniblocks, err := s.txCoordinator.ProcessBlockTransaction(headerHandler, blockBody, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, nil, err
	}

	postProcessMBs := s.txCoordinator.CreatePostProcessMiniBlocks()

	receiptsHash, err := s.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, nil, err
	}

	err = headerHandler.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, nil, err
	}

	s.prepareBlockHeaderInternalMapForValidatorProcessor()
	_, err = s.validatorStatisticsProcessor.UpdatePeerState(headerHandler, makeCommonHeaderHandlerHashMap(s.hdrsForCurrBlock.getHdrHashMap()))
	if err != nil {
		return nil, nil, err
	}

	createdBlockBody := &block.Body{MiniBlocks: miniblocks}
	createdBlockBody.MiniBlocks = append(createdBlockBody.MiniBlocks, postProcessMBs...)
	newBody, err := s.applyBodyToHeader(headerHandler, createdBlockBody)
	if err != nil {
		return nil, nil, err
	}

	return headerHandler, newBody, nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (s *sovereignBlockProcessor) applyBodyToHeader(
	headerHandler data.HeaderHandler,
	body *block.Body,
) (*block.Body, error) {
	sw := core.NewStopWatch()
	sw.Start("applyBodyToHeader")
	defer func() {
		sw.Stop("applyBodyToHeader")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	var err error
	err = headerHandler.SetMiniBlockHeaderHandlers(nil)
	if err != nil {
		return nil, err
	}

	rootHash, err := s.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetRootHash(rootHash)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := s.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetValidatorStatsRootHash(validatorStatsRootHash)
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)
	err = s.applyBodyInfoOnCommonHeader(headerHandler, newBody, nil)
	if err != nil {
		return nil, err
	}
	return newBody, nil
}

// TODO: verify if block created from processblock is the same one as received from leader - without signature - no need for another set of checks
// actually sign check should resolve this - as you signed something generated by you

// CommitBlock - will do a lot of verification
func (s *sovereignBlockProcessor) CommitBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	s.processStatusHandler.SetBusy("sovereignBlockProcessor.CommitBlock")
	var err error
	defer func() {
		if err != nil {
			s.RevertCurrentBlock()
		}
		s.processStatusHandler.SetIdle()
	}()

	err = checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = s.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	err = s.verifyFees(headerHandler)
	if err != nil {
		return err
	}

	if !s.verifyStateRoot(headerHandler.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = s.verifyValidatorStatisticsRootHash(headerHandler)
	if err != nil {
		return err
	}

	marshalizedHeader, err := s.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	headerHash := s.hasher.Compute(string(marshalizedHeader))
	s.saveShardHeader(headerHandler, headerHash, marshalizedHeader)
	s.saveBody(body, headerHandler, headerHash)

	err = s.commitAll(headerHandler)
	if err != nil {
		return err
	}

	s.validatorStatisticsProcessor.DisplayRatings(headerHandler.GetEpoch())

	err = s.forkDetector.AddHeader(headerHandler, headerHash, process.BHProcessed, nil, nil)
	if err != nil {
		log.Debug("forkDetector.AddHeader", "error", err.Error())
		return err
	}

	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := getLastSelfNotarizedHeaderByItself(s.blockChain)
	s.blockTracker.AddSelfNotarizedHeader(s.shardCoordinator.SelfId(), lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	go s.historyRepo.OnNotarizedBlocks(s.shardCoordinator.SelfId(), []data.HeaderHandler{lastSelfNotarizedHeader}, [][]byte{lastSelfNotarizedHeaderHash})

	s.updateState(lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	highestFinalBlockNonce := s.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"shard", s.shardCoordinator.SelfId(),
		"nonce", highestFinalBlockNonce,
	)

	err = s.commonHeaderAndBodyCommit(headerHandler, body, headerHash, []data.HeaderHandler{lastSelfNotarizedHeader}, [][]byte{lastSelfNotarizedHeaderHash})
	if err != nil {
		return err
	}

	return nil
}

// RestoreBlockIntoPools restores block into pools
func (s *sovereignBlockProcessor) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	s.restoreBlockBody(header, body)
	s.blockTracker.RemoveLastNotarizedHeaders()
	return nil
}

// RevertStateToBlock reverts state in tries
func (s *sovereignBlockProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	return s.revertAccountsStates(header, rootHash)
}

// RevertCurrentBlock reverts the current block for cleanup failed process
func (s *sovereignBlockProcessor) RevertCurrentBlock() {
	s.revertAccountState()
}

// PruneStateOnRollback prunes states of all accounts DBs
func (s *sovereignBlockProcessor) PruneStateOnRollback(currHeader data.HeaderHandler, _ []byte, prevHeader data.HeaderHandler, _ []byte) {
	for key := range s.accountsDB {
		if !s.accountsDB[key].IsPruningEnabled() {
			continue
		}

		rootHash, prevRootHash := s.getRootHashes(currHeader, prevHeader, key)
		if bytes.Equal(rootHash, prevRootHash) {
			continue
		}

		s.accountsDB[key].CancelPrune(prevRootHash, state.OldRoot)
		s.accountsDB[key].PruneTrie(rootHash, state.NewRoot, s.getPruningHandler(currHeader.GetNonce()))
	}
}

// ProcessScheduledBlock does nothing - as this uses new execution model
func (s *sovereignBlockProcessor) ProcessScheduledBlock(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
	return nil
}

// DecodeBlockHeader decodes the current header
func (s *sovereignBlockProcessor) DecodeBlockHeader(data []byte) data.HeaderHandler {
	if data == nil {
		return nil
	}

	header, err := process.UnmarshalSovereignChainHeader(s.marshalizer, data)
	if err != nil {
		log.Debug("DecodeBlockHeader.UnmarshalSovereignChainHeader", "error", err.Error())
		return nil
	}

	return header
}

func (s *sovereignBlockProcessor) verifyValidatorStatisticsRootHash(headerHandler data.HeaderHandler) error {
	validatorStatsRH, err := s.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return err
	}

	if !bytes.Equal(validatorStatsRH, headerHandler.GetValidatorStatsRootHash()) {
		log.Debug("validator stats root hash mismatch",
			"computed", validatorStatsRH,
			"received", headerHandler.GetValidatorStatsRootHash(),
		)
		return fmt.Errorf("%s, sovereign, computed: %s, received: %s, header nonce: %d",
			process.ErrValidatorStatsRootHashDoesNotMatch,
			logger.DisplayByteSlice(validatorStatsRH),
			logger.DisplayByteSlice(headerHandler.GetValidatorStatsRootHash()),
			headerHandler.GetNonce(),
		)
	}

	return nil
}

func (s *sovereignBlockProcessor) updateState(header data.HeaderHandler, headerHash []byte) {
	if check.IfNil(header) {
		log.Debug("updateState nil header")
		return
	}

	s.validatorStatisticsProcessor.SetLastFinalizedRootHash(header.GetValidatorStatsRootHash())

	prevHeaderHash := header.GetPrevHash()
	prevHeader, errNotCritical := process.GetSovereignChainHeader(
		prevHeaderHash,
		s.dataPool.Headers(),
		s.marshalizer,
		s.store,
	)
	if errNotCritical != nil {
		log.Debug("could not get header with validator stats from storage")
		return
	}

	s.updateStateStorage(
		header,
		header.GetRootHash(),
		prevHeader.GetRootHash(),
		s.accountsDB[state.UserAccountsState],
	)

	s.updateStateStorage(
		header,
		header.GetValidatorStatsRootHash(),
		prevHeader.GetValidatorStatsRootHash(),
		s.accountsDB[state.PeerAccountsState],
	)

	s.setFinalizedHeaderHashInIndexer(header.GetPrevHash())
	s.blockChain.SetFinalBlockInfo(header.GetNonce(), headerHash, header.GetRootHash())
}

// IsInterfaceNil returns true if underlying object is nil
func (s *sovereignBlockProcessor) IsInterfaceNil() bool {
	return s == nil
}
