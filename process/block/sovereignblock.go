package block

import (
	"bytes"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/state"
	"math/big"
	"time"
)

type ArgsSovereignBlockProcessor struct {
	ArgShardProcessor
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
}

type sovereignBlockProcessor struct {
	*shardProcessor
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
}

// NewSovereignBlockProcessor creates a new sovereign block processor
func NewSovereignBlockProcessor(arguments ArgsSovereignBlockProcessor) (*sovereignBlockProcessor, error) {
	sp, err := NewShardProcessor(arguments.ArgShardProcessor)
	if err != nil {
		return nil, err
	}

	sovereign := &sovereignBlockProcessor{
		shardProcessor:               sp,
		validatorStatisticsProcessor: arguments.ValidatorStatisticsProcessor,
	}

	sovereign.scheduledTxsExecutionHandler = &disabled.ScheduledTxsExecutionHandler{}

	return sovereign, nil
}

// CreateNewHeader creates a new header
func (s *sovereignBlockProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	s.enableRoundsHandler.CheckRound(round)
	header := &block.HeaderWithValidatorStats{
		Header: &block.Header{
			SoftwareVersion: []byte("1"),
		},
	}

	err := header.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = header.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = header.SetAccumulatedFees(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = header.SetDeveloperFees(big.NewInt(0))
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
	commonHdr, ok := initialHdr.(data.CommonHeaderHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	s.processStatusHandler.SetBusy("shardProcessor.CreateBlock")
	defer s.processStatusHandler.SetIdle()

	for _, accounts := range s.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("metaProcessor.CreateBlock first entry", "stack", accounts.GetStackDebugFirstEntry()))
			return nil, nil, process.ErrAccountStateDirty
		}
	}

	err := s.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	s.blockChainHook.SetCurrentHeader(commonHdr)
	startTime := time.Now()
	mbsFromMe := s.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, commonHdr.GetPrevRandSeed())
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs from me", "time", elapsedTime)

	sw.Start("UpdatePeerState")
	mp.prepareBlockHeaderInternalMapForValidatorProcessor()
	valStatRootHash, err := mp.validatorStatisticsProcessor.UpdatePeerState(metaHdr, makeCommonHeaderHandlerHashMap(mp.hdrsForCurrBlock.getHdrHashMap()))
	sw.Stop("UpdatePeerState")
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetValidatorStatsRootHash(valStatRootHash)
	if err != nil {
		return nil, err
	}

	return initialHdr, &block.Body{MiniBlocks: mbsFromMe}, nil
}

// ProcessBlock actually process the selected transaction and will create the final block body
func (s *sovereignBlockProcessor) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	s.processStatusHandler.SetBusy("shardProcessor.ProcessBlock")
	defer s.processStatusHandler.SetIdle()

	err := s.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			go s.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	sovereignHdr, ok := header.(*block.HeaderWithValidatorStats)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	blockBody, ok := body.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = s.createBlockStarted()
	if err != nil {
		return err
	}

	s.txCoordinator.RequestBlockTransactions(body)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = s.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	for _, accounts := range s.accountsDB {
		if accounts.JournalLen() != 0 {
			log.Error("metaProcessor.CreateBlock first entry", "stack", accounts.GetStackDebugFirstEntry()))
			return process.ErrAccountStateDirty
		}
	}

	defer func() {
		if err != nil {
			s.RevertCurrentBlock()
		}
	}()

	startTime := time.Now()
	err = s.txCoordinator.ProcessBlockTransaction(sovereignHdr, blockBody, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	s.prepareBlockHeaderInternalMapForValidatorProcessor()
	validatorStatsRH, err := s.validatorStatisticsProcessor.UpdatePeerState(header, makeCommonHeaderHandlerHashMap(s.hdrsForCurrBlock.getHdrHashMap()))
	if err != nil {
		return err
	}

	err = sovereignHdr.SetValidatorStatsRootHash(validatorStatsRH)
	if err != nil {
		return err
	}

	finalBody, err := s.applyBodyToHeader(sovereignHdr, blockBody)
	if err != nil {
		return err
	}
	blockBody = finalBody

	return nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (s* sovereignBlockProcessor) applyBodyToHeader(
	sovereignHdr *block.HeaderWithValidatorStats,
	body *block.Body,
) (*block.Body, error) {
	sw := core.NewStopWatch()
	sw.Start("applyBodyToHeader")
	defer func() {
		sw.Stop("applyBodyToHeader")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	var err error
	err = sovereignHdr.SetMiniBlockHeaderHandlers(nil)
	if err != nil {
		return nil, err
	}

	rootHash, err := s.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = sovereignHdr.SetRootHash(rootHash)
	if err != nil {
		return nil, err
	}

	rootHash, err = s.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return nil, err
	}
	err = sovereignHdr.SetValidatorStatsRootHash(rootHash)
	if err != nil {
		return nil, err
	}

	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	var receiptsHash []byte
	sw.Start("CreateReceiptsHash")
	receiptsHash, err = s.txCoordinator.CreateReceiptsHash()
	sw.Stop("CreateReceiptsHash")
	if err != nil {
		return nil, err
	}

	err = sovereignHdr.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)

	sw.Start("createMiniBlockHeaders")
	totalTxCount, miniBlockHeaderHandlers, err := s.createMiniBlockHeaderHandlers(newBody, nil)
	sw.Stop("createMiniBlockHeaders")
	if err != nil {
		return nil, err
	}

	err = sovereignHdr.SetMiniBlockHeaderHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

	err = sovereignHdr.SetTxCount(uint32(totalTxCount))
	if err != nil {
		return nil, err
	}

	err = sovereignHdr.SetAccumulatedFees(s.feeHandler.GetAccumulatedFees())
	if err != nil {
		return nil, err
	}

	err = sovereignHdr.SetDeveloperFees(s.feeHandler.GetDeveloperFees())
	if err != nil {
		return nil, err
	}

	err = s.txCoordinator.VerifyCreatedMiniBlocks(sovereignHdr, newBody)
	if err != nil {
		return nil, err
	}

	s.appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(totalTxCount))
	s.appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, uint64(len(body.MiniBlocks)))

	marshaledBody, err := s.marshalizer.Marshal(newBody)
	if err != nil {
		return nil, err
	}
	s.blockSizeThrottler.Add(sovereignHdr.GetRound(), uint32(len(marshaledBody)))

	return newBody, nil
}

// TODO: verify if block created from processblock is the same one as received from leader - without signature - no need for another set of checks
// actually sign check should resolve this - as you signed something generated by you

// CommitBlock - will do a lot of verification
func (s *sovereignBlockProcessor) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	var err error
	err = s.verifyFees(header)
	if err != nil {
		return err
	}

	if !s.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = s.verifyValidatorStatisticsRootHash(header)
	if err != nil {
		return err
	}

	return nil
}

// RestoreBlockIntoPools restore block into pools
func (s *sovereignBlockProcessor) RestoreBlockIntoPools(_ data.HeaderHandler, body data.BodyHandler) error {
	s.restoreBlockBody(body)
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

// PruneStateOnRollback prune states of all accounts DBs
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
func (s *sovereignBlockProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	header, err := process.UnmarshalHeaderWithValidatorStats(s.marshalizer, dta)
	if err != nil {
		log.Debug("DecodeBlockHeader.UnmarshalShardHeader", "error", err.Error())
		return nil
	}

	return header
}

func (s *sovereignBlockProcessor) verifyValidatorStatisticsRootHash(header data.CommonHeaderHandler) error {
	validatorStatsRH, err := s.accountsDB[state.PeerAccountsState].RootHash()
	if err != nil {
		return err
	}

	validatorStatsInfo, ok := header.(data.ValidatorStatisticsInfoHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if !bytes.Equal(validatorStatsRH, validatorStatsInfo.GetValidatorStatsRootHash()) {
		log.Debug("validator stats root hash mismatch",
			"computed", validatorStatsRH,
			"received", validatorStatsInfo.GetValidatorStatsRootHash(),
		)
		return fmt.Errorf("%s, metachain, computed: %s, received: %s, meta header nonce: %d",
			process.ErrValidatorStatsRootHashDoesNotMatch,
			logger.DisplayByteSlice(validatorStatsRH),
			logger.DisplayByteSlice(validatorStatsInfo.GetValidatorStatsRootHash()),
			header.GetNonce(),
		)
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *sovereignBlockProcessor) IsInterfaceNil() bool {
	return s == nil
}
