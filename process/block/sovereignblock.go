package block

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
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

	return sovereign, nil
}

func (s *sovereignBlockProcessor) ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	//TODO implement me
	panic("implement me")
}

// ProcessScheduledBlock does nothing - as this uses new execution model
func (s *sovereignBlockProcessor) ProcessScheduledBlock(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
	return nil
}

func (s *sovereignBlockProcessor) CommitBlock(header data.HeaderHandler, body data.BodyHandler) error {
	//TODO implement me
	panic("implement me")
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

// RevertStateToBlock reverts state in tries
func (s *sovereignBlockProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	return s.revertAccountsStates(header, rootHash)
}

func (s *sovereignBlockProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	//TODO implement me
	panic("implement me")
}

func (s *sovereignBlockProcessor) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	//TODO implement me
	panic("implement me")
}

func (s *sovereignBlockProcessor) CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
	//TODO implement me
	panic("implement me")
}

func (s *sovereignBlockProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	//TODO implement me
	panic("implement me")
}

// IsInterfaceNil returns true if underlying object is nil
func (s *sovereignBlockProcessor) IsInterfaceNil() bool {
	return s == nil
}
