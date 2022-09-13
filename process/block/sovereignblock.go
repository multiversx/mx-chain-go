package block

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
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
	shardHdr, ok := initialHdr.(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	s.processStatusHandler.SetBusy("shardProcessor.CreateBlock")
	defer s.processStatusHandler.SetIdle()

	err := s.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	s.blockChainHook.SetCurrentHeader(shardHdr)
	body, processedMiniBlocksDestMeInfo, err := sp.createBlockBody(shardHdr, haveTime)
	if err != nil {
		return nil, nil, err
	}

	finalBody, err := sp.applyBodyToHeader(shardHdr, body, processedMiniBlocksDestMeInfo)
	if err != nil {
		return nil, nil, err
	}

	return shardHdr, finalBody, nil
}

// ProcessBlock actually process the selected transaction and will create the final block body
func (s *sovereignBlockProcessor) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	//TODO implement me
	panic("implement me")
}

// CommitBlock - will do a lot of verification
func (s *sovereignBlockProcessor) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	//TODO implement me
	panic("implement me")
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

// IsInterfaceNil returns true if underlying object is nil
func (s *sovereignBlockProcessor) IsInterfaceNil() bool {
	return s == nil
}
