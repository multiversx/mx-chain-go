package block

import (
	"bytes"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
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
		return nil, process.ErrWrongTypeAssertion
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
	startTime := time.Now()
	mbsFromMe := s.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, initialHdr.GetPrevRandSeed())
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs from me", "time", elapsedTime)

	return initialHdr, &block.Body{MiniBlocks: mbsFromMe}, nil
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
