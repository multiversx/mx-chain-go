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
	logger "github.com/multiversx/mx-chain-logger-go"
)

var rootHash = "uncomputed root hash"

type extendedShardHeaderTrackHandler interface {
	ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error)
}

type sovereignChainBlockProcessor struct {
	*shardProcessor
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	uncomputedRootHash           []byte
	extendedShardHeaderTracker   extendedShardHeaderTrackHandler
}

// NewSovereignChainBlockProcessor creates a new sovereign chain block processor
func NewSovereignChainBlockProcessor(
	shardProcessor *shardProcessor,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
) (*sovereignChainBlockProcessor, error) {
	if shardProcessor == nil {
		return nil, process.ErrNilBlockProcessor
	}
	if validatorStatisticsProcessor == nil {
		return nil, process.ErrNilValidatorStatistics
	}

	scbp := &sovereignChainBlockProcessor{
		shardProcessor:               shardProcessor,
		validatorStatisticsProcessor: validatorStatisticsProcessor,
	}

	scbp.uncomputedRootHash = scbp.hasher.Compute(rootHash)

	extendedShardHeaderTracker, ok := scbp.blockTracker.(extendedShardHeaderTrackHandler)
	if !ok {
		return nil, fmt.Errorf("%w in NewSovereignBlockProcessor", process.ErrWrongTypeAssertion)
	}

	scbp.extendedShardHeaderTracker = extendedShardHeaderTracker

	//TODO: This call and the method itself should be removed when real functionality will be done
	scbp.addNextTrackedHeadersMock(3)

	return scbp, nil
}

func (scbp *sovereignChainBlockProcessor) addNextTrackedHeadersMock(numHeadersToBeAdded int) {
	lastHeader, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
	if err != nil {
		log.Debug("sovereignChainBlockProcessor.addNextTrackedHeaderMock", "error", err.Error())
		return
	}
	if check.IfNil(lastHeader) {
		log.Debug("sovereignChainBlockProcessor.addNextTrackedHeaderMock", "error", process.ErrNilHeaderHandler)
		return
	}

	shardHeaderExtended, isShardHeaderExtended := lastHeader.(*block.ShardHeaderExtended)
	if !isShardHeaderExtended {
		log.Debug("sovereignChainBlockProcessor.addNextTrackedHeaderMock", "error", process.ErrWrongTypeAssertion)
		return
	}

	lastHeaderHash, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, shardHeaderExtended.Header.Header)

	for i := 0; i < numHeadersToBeAdded; i++ {
		randSeed, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, &block.Header{Reserved: []byte(time.Now().String())})

		time.Sleep(10 * time.Millisecond)
		txHash1, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, &block.Header{Reserved: []byte(time.Now().String())})

		time.Sleep(10 * time.Millisecond)
		txHash2, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, &block.Header{Reserved: []byte(time.Now().String())})

		time.Sleep(10 * time.Millisecond)
		txHash3, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, &block.Header{Reserved: []byte(time.Now().String())})

		incomingTxHashes := [][]byte{
			txHash1,
			txHash2,
			txHash3,
		}

		incomingMiniBlocks := []*block.MiniBlock{
			{
				Type:            block.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: core.SovereignChainShardId,
				TxHashes:        incomingTxHashes,
			},
		}

		header := &block.Header{
			Nonce:        lastHeader.GetNonce() + 1,
			Round:        lastHeader.GetRound() + 1,
			PrevHash:     lastHeaderHash,
			PrevRandSeed: lastHeader.GetRandSeed(),
			RandSeed:     randSeed,
		}

		nextCrossNotarizedHeader := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: header,
			},
			IncomingMiniBlocks: incomingMiniBlocks,
		}
		nextCrossNotarizedHeaderHash, _ := core.CalculateHash(scbp.marshalizer, scbp.hasher, nextCrossNotarizedHeader)

		scbp.blockTracker.AddTrackedHeader(nextCrossNotarizedHeader, nextCrossNotarizedHeaderHash)

		lastHeader = header
		lastHeaderHash, _ = core.CalculateHash(scbp.marshalizer, scbp.hasher, header)
	}
}

// CreateNewHeader creates a new header
func (scbp *sovereignChainBlockProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	scbp.enableRoundsHandler.CheckRound(round)
	header := &block.SovereignChainHeader{
		Header: &block.Header{
			SoftwareVersion: process.SovereignHeaderVersion,
			RootHash:        scbp.uncomputedRootHash,
		},
		ValidatorStatsRootHash: scbp.uncomputedRootHash,
	}

	err := scbp.setRoundNonceInitFees(round, nonce, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// CreateBlock selects and puts transaction into the temporary block body
func (scbp *sovereignChainBlockProcessor) CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}

	sovereignChainHeaderHandler, ok := initialHdr.(data.SovereignChainHeaderHandler)
	if !ok {
		return nil, nil, fmt.Errorf("%w in sovereignChainBlockProcessor.CreateBlock", process.ErrWrongTypeAssertion)
	}

	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.CreateBlock")
	defer scbp.processStatusHandler.SetIdle()

	for _, accounts := range scbp.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignChainBlockProcessor.CreateBlock first entry", "stack", accounts.GetStackDebugFirstEntry())
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	err := scbp.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	scbp.blockChainHook.SetCurrentHeader(initialHdr)

	var miniBlocks block.MiniBlockSlice
	//processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	if !haveTime() {
		log.Debug("sovereignChainBlockProcessor.CreateBlock", "error", process.ErrTimeIsOut)

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return nil, nil, process.ErrTimeIsOut
	}

	startTime := time.Now()
	createIncomingMiniBlocksDestMeInfo, err := scbp.createIncomingMiniBlocksDestMe(haveTime)
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
	mbsFromMe := scbp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, initialHdr.GetPrevRandSeed())
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

	mainChainShardHeaderHashes := scbp.sortExtendedShardHeaderHashesForCurrentBlockByNonce(true)
	err = sovereignChainHeaderHandler.SetMainChainShardHeaderHashes(mainChainShardHeaderHashes)
	if err != nil {
		return nil, nil, err
	}

	return initialHdr, &block.Body{MiniBlocks: miniBlocks}, nil
}

func (scbp *sovereignChainBlockProcessor) createIncomingMiniBlocksDestMe(haveTime func() bool) (*createAndProcessMiniBlocksDestMeInfo, error) {
	log.Debug("createIncomingMiniBlocksDestMe has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestExtendedShardChainFromLastNotarized")
	orderedExtendedShardHeaders, orderedExtendedShardHeadersHashes, err := scbp.extendedShardHeaderTracker.ComputeLongestExtendedShardChainFromLastNotarized()
	sw.Stop("ComputeLongestExtendedShardChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return nil, err
	}

	log.Debug("extended shard headers ordered",
		"num extended shard headers", len(orderedExtendedShardHeaders),
	)

	lastExtendedShardHdr, _, err := scbp.blockTracker.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
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
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
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
			scbp.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currHdrHash)] = &hdrInfo{hdr: createAndProcessInfo.currHdr, usedInBlock: true}
			createAndProcessInfo.numHdrsAdded++
			lastExtendedShardHdr = createAndProcessInfo.currHdr
			continue
		}

		createAndProcessInfo.currProcessedMiniBlocksInfo = scbp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(createAndProcessInfo.currHdrHash)
		createAndProcessInfo.hdrAdded = false

		shouldContinue, errCreated := scbp.createIncomingMiniBlocksAndTransactionsDestMe(createAndProcessInfo)
		if errCreated != nil {
			return nil, errCreated
		}
		if !shouldContinue {
			break
		}

		lastExtendedShardHdr = createAndProcessInfo.currHdr
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	go scbp.requestExtendedShardHeadersIfNeeded(createAndProcessInfo.numHdrsAdded, lastExtendedShardHdr)

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

func (scbp *sovereignChainBlockProcessor) createIncomingMiniBlocksAndTransactionsDestMe(
	createAndProcessInfo *createAndProcessMiniBlocksDestMeInfo,
) (bool, error) {
	currMiniBlocksAdded, currNumTxsAdded, hdrProcessFinished, errCreated := scbp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
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
		scbp.hdrsForCurrBlock.hdrHashAndInfo[string(createAndProcessInfo.currHdrHash)] = &hdrInfo{hdr: createAndProcessInfo.currHdr, usedInBlock: true}
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

func (scbp *sovereignChainBlockProcessor) requestExtendedShardHeadersIfNeeded(hdrsAdded uint32, lastExtendedShardHdr data.HeaderHandler) {
	log.Debug("extended shard headers added",
		"num", hdrsAdded,
		"highest nonce", lastExtendedShardHdr.GetNonce(),
	)
	//TODO: A request mechanism should be implemented if extended shard header(s) is(are) needed
}

func (scbp *sovereignChainBlockProcessor) sortExtendedShardHeaderHashesForCurrentBlockByNonce(usedInBlock bool) [][]byte {
	hdrsForCurrentBlockInfo := make([]*nonceAndHashInfo, 0)

	scbp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for headerHash, headerInfo := range scbp.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlockInfo = append(hdrsForCurrentBlockInfo,
			&nonceAndHashInfo{nonce: headerInfo.hdr.GetNonce(), hash: []byte(headerHash)})
	}
	scbp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

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
func (scbp *sovereignChainBlockProcessor) ProcessBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
	if haveTime == nil {
		return nil, nil, process.ErrNilHaveTimeHandler
	}

	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.ProcessBlock")
	defer scbp.processStatusHandler.SetIdle()

	err := scbp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			go scbp.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return nil, nil, err
	}

	blockBody, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	err = scbp.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	scbp.txCoordinator.RequestBlockTransactions(blockBody)
	err = scbp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return nil, nil, err
	}

	for _, accounts := range scbp.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("sovereignChainBlockProcessor.ProcessBlock first entry", "stack", accounts.GetStackDebugFirstEntry())
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	defer func() {
		if err != nil {
			scbp.RevertCurrentBlock()
		}
	}()

	startTime := time.Now()
	miniblocks, err := scbp.txCoordinator.ProcessBlockTransaction(headerHandler, blockBody, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, nil, err
	}

	postProcessMBs := scbp.txCoordinator.CreatePostProcessMiniBlocks()

	receiptsHash, err := scbp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, nil, err
	}

	err = headerHandler.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, nil, err
	}

	scbp.prepareBlockHeaderInternalMapForValidatorProcessor()
	_, err = scbp.validatorStatisticsProcessor.UpdatePeerState(headerHandler, makeCommonHeaderHandlerHashMap(scbp.hdrsForCurrBlock.getHdrHashMap()))
	if err != nil {
		return nil, nil, err
	}

	createdBlockBody := &block.Body{MiniBlocks: miniblocks}
	createdBlockBody.MiniBlocks = append(createdBlockBody.MiniBlocks, postProcessMBs...)
	newBody, err := scbp.applyBodyToHeader(headerHandler, createdBlockBody)
	if err != nil {
		return nil, nil, err
	}

	return headerHandler, newBody, nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (scbp *sovereignChainBlockProcessor) applyBodyToHeader(
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

	rootHash, err := scbp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetRootHash(rootHash)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := scbp.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = headerHandler.SetValidatorStatsRootHash(validatorStatsRootHash)
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)
	err = scbp.applyBodyInfoOnCommonHeader(headerHandler, newBody, nil)
	if err != nil {
		return nil, err
	}
	return newBody, nil
}

// TODO: verify if block created from processblock is the same one as received from leader - without signature - no need for another set of checks
// actually sign check should resolve this - as you signed something generated by you

// CommitBlock - will do a lot of verification
func (scbp *sovereignChainBlockProcessor) CommitBlock(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	scbp.processStatusHandler.SetBusy("sovereignChainBlockProcessor.CommitBlock")
	var err error
	defer func() {
		if err != nil {
			scbp.RevertCurrentBlock()
		}
		scbp.processStatusHandler.SetIdle()
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

	err = scbp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	err = scbp.verifyFees(headerHandler)
	if err != nil {
		return err
	}

	if !scbp.verifyStateRoot(headerHandler.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = scbp.verifyValidatorStatisticsRootHash(headerHandler)
	if err != nil {
		return err
	}

	marshalizedHeader, err := scbp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	headerHash := scbp.hasher.Compute(string(marshalizedHeader))
	scbp.saveShardHeader(headerHandler, headerHash, marshalizedHeader)
	scbp.saveBody(body, headerHandler, headerHash)

	err = scbp.commitAll(headerHandler)
	if err != nil {
		return err
	}

	scbp.validatorStatisticsProcessor.DisplayRatings(headerHandler.GetEpoch())

	err = scbp.forkDetector.AddHeader(headerHandler, headerHash, process.BHProcessed, nil, nil)
	if err != nil {
		log.Debug("forkDetector.AddHeader", "error", err.Error())
		return err
	}

	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := getLastSelfNotarizedHeaderByItself(scbp.blockChain)
	scbp.blockTracker.AddSelfNotarizedHeader(scbp.shardCoordinator.SelfId(), lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	go scbp.historyRepo.OnNotarizedBlocks(scbp.shardCoordinator.SelfId(), []data.HeaderHandler{lastSelfNotarizedHeader}, [][]byte{lastSelfNotarizedHeaderHash})

	scbp.updateState(lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	highestFinalBlockNonce := scbp.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"shard", scbp.shardCoordinator.SelfId(),
		"nonce", highestFinalBlockNonce,
	)

	err = scbp.commonHeaderAndBodyCommit(headerHandler, body, headerHash, []data.HeaderHandler{lastSelfNotarizedHeader}, [][]byte{lastSelfNotarizedHeaderHash})
	if err != nil {
		return err
	}

	return nil
}

// RestoreBlockIntoPools restores block into pools
func (scbp *sovereignChainBlockProcessor) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	scbp.restoreBlockBody(header, body)
	scbp.blockTracker.RemoveLastNotarizedHeaders()
	return nil
}

// RevertStateToBlock reverts state in tries
func (scbp *sovereignChainBlockProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	return scbp.revertAccountsStates(header, rootHash)
}

// RevertCurrentBlock reverts the current block for cleanup failed process
func (scbp *sovereignChainBlockProcessor) RevertCurrentBlock() {
	scbp.revertAccountState()
}

// PruneStateOnRollback prunes states of all accounts DBs
func (scbp *sovereignChainBlockProcessor) PruneStateOnRollback(currHeader data.HeaderHandler, _ []byte, prevHeader data.HeaderHandler, _ []byte) {
	for key := range scbp.accountsDB {
		if !scbp.accountsDB[key].IsPruningEnabled() {
			continue
		}

		rootHash, prevRootHash := scbp.getRootHashes(currHeader, prevHeader, key)
		if bytes.Equal(rootHash, prevRootHash) {
			continue
		}

		scbp.accountsDB[key].CancelPrune(prevRootHash, state.OldRoot)
		scbp.accountsDB[key].PruneTrie(rootHash, state.NewRoot, scbp.getPruningHandler(currHeader.GetNonce()))
	}
}

// ProcessScheduledBlock does nothing - as this uses new execution model
func (scbp *sovereignChainBlockProcessor) ProcessScheduledBlock(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
	return nil
}

// DecodeBlockHeader decodes the current header
func (scbp *sovereignChainBlockProcessor) DecodeBlockHeader(data []byte) data.HeaderHandler {
	if data == nil {
		return nil
	}

	header, err := process.UnmarshalSovereignChainHeader(scbp.marshalizer, data)
	if err != nil {
		log.Debug("DecodeBlockHeader.UnmarshalSovereignChainHeader", "error", err.Error())
		return nil
	}

	return header
}

func (scbp *sovereignChainBlockProcessor) verifyValidatorStatisticsRootHash(headerHandler data.HeaderHandler) error {
	validatorStatsRH, err := scbp.accountsDB[state.PeerAccountsState].RootHash()
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

func (scbp *sovereignChainBlockProcessor) updateState(header data.HeaderHandler, headerHash []byte) {
	if check.IfNil(header) {
		log.Debug("updateState nil header")
		return
	}

	scbp.validatorStatisticsProcessor.SetLastFinalizedRootHash(header.GetValidatorStatsRootHash())

	prevHeaderHash := header.GetPrevHash()
	prevHeader, errNotCritical := process.GetSovereignChainHeader(
		prevHeaderHash,
		scbp.dataPool.Headers(),
		scbp.marshalizer,
		scbp.store,
	)
	if errNotCritical != nil {
		log.Debug("could not get header with validator stats from storage")
		return
	}

	scbp.updateStateStorage(
		header,
		header.GetRootHash(),
		prevHeader.GetRootHash(),
		scbp.accountsDB[state.UserAccountsState],
	)

	scbp.updateStateStorage(
		header,
		header.GetValidatorStatsRootHash(),
		prevHeader.GetValidatorStatsRootHash(),
		scbp.accountsDB[state.PeerAccountsState],
	)

	scbp.setFinalizedHeaderHashInIndexer(header.GetPrevHash())
	scbp.blockChain.SetFinalBlockInfo(header.GetNonce(), headerHash, header.GetRootHash())
}

// IsInterfaceNil returns true if underlying object is nil
func (scbp *sovereignChainBlockProcessor) IsInterfaceNil() bool {
	return scbp == nil
}
