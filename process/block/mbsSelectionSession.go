package block

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/process"
)

type miniBlocksSelectionConfig struct {
	shardID    uint32
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
}

type miniBlocksSelectionResult struct {
	miniBlockHeaderHandlers   []data.MiniBlockHeaderHandler
	miniBlocks                block.MiniBlockSlice
	miniBlockHashes           [][]byte
	referencedMetaBlockHashes [][]byte
	referencedMetaBlocks      []data.HeaderHandler
	lastMetaBlock             data.HeaderHandler
	gasProvided               uint64
	numTxsAdded               uint32
	mut                       sync.RWMutex
}

type miniBlocksSelectionSession struct {
	*miniBlocksSelectionConfig
	*miniBlocksSelectionResult
}

const defaultCapacity = 10

func NewMiniBlocksSelectionSession(shardID uint32, marshaller marshal.Marshalizer, hasher hashing.Hasher) (*miniBlocksSelectionSession, error) {
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}

	return &miniBlocksSelectionSession{
		miniBlocksSelectionConfig: &miniBlocksSelectionConfig{
			shardID:    shardID,
			marshaller: marshaller,
			hasher:     hasher,
		},
		miniBlocksSelectionResult: &miniBlocksSelectionResult{
			miniBlockHeaderHandlers:   make([]data.MiniBlockHeaderHandler, 0, defaultCapacity),
			miniBlocks:                make(block.MiniBlockSlice, 0, defaultCapacity),
			miniBlockHashes:           make([][]byte, 0, defaultCapacity),
			referencedMetaBlockHashes: make([][]byte, 0, defaultCapacity),
			referencedMetaBlocks:      make([]data.HeaderHandler, 0, defaultCapacity),
			lastMetaBlock:             nil,
			gasProvided:               0,
			numTxsAdded:               0,
		},
	}, nil
}

// ResetSelectionSession resets the mini blocks selection session by clearing all the fields
func (s *miniBlocksSelectionSession) ResetSelectionSession() {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.miniBlockHeaderHandlers = make([]data.MiniBlockHeaderHandler, 0, defaultCapacity)
	s.miniBlocks = make(block.MiniBlockSlice, 0, defaultCapacity)
	s.miniBlockHashes = make([][]byte, 0, defaultCapacity)
	s.referencedMetaBlockHashes = make([][]byte, 0, defaultCapacity)
	s.referencedMetaBlocks = make([]data.HeaderHandler, 0, defaultCapacity)
	s.lastMetaBlock = nil
	s.gasProvided = 0
	s.numTxsAdded = 0
}

// GetMiniBlockHeaderHandlers returns the mini block header handlers
func (s *miniBlocksSelectionSession) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.miniBlockHeaderHandlers) == 0 {
		return nil
	}

	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(s.miniBlockHeaderHandlers))
	copy(miniBlockHeaderHandlers, s.miniBlockHeaderHandlers)

	return miniBlockHeaderHandlers
}

// GetMiniBlocks returns the mini blocks
func (s *miniBlocksSelectionSession) GetMiniBlocks() block.MiniBlockSlice {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.miniBlocks) == 0 {
		return nil
	}

	miniBlocks := make(block.MiniBlockSlice, len(s.miniBlocks))
	copy(miniBlocks, s.miniBlocks)

	return miniBlocks
}

// GetMiniBlockHashes returns the hashes of the mini blocks
func (s *miniBlocksSelectionSession) GetMiniBlockHashes() [][]byte {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.miniBlockHashes) == 0 {
		return nil
	}

	miniBlockHashes := make([][]byte, len(s.miniBlockHashes))
	copy(miniBlockHashes, s.miniBlockHashes)

	return miniBlockHashes
}

// AddReferencedMetaBlock adds a meta block and its hash to the session
func (s *miniBlocksSelectionSession) AddReferencedMetaBlock(metaBlock data.HeaderHandler, metaBlockHash []byte) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if check.IfNil(metaBlock) || len(metaBlockHash) == 0 {
		return
	}

	s.referencedMetaBlocks = append(s.referencedMetaBlocks, metaBlock)
	s.referencedMetaBlockHashes = append(s.referencedMetaBlockHashes, metaBlockHash)
	s.lastMetaBlock = metaBlock
}

// GetReferencedMetaBlockHashes returns the hashes of the referenced meta blocks
func (s *miniBlocksSelectionSession) GetReferencedMetaBlockHashes() [][]byte {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.referencedMetaBlockHashes) == 0 {
		return nil
	}

	referencedMetaBlockHashes := make([][]byte, len(s.referencedMetaBlockHashes))
	copy(referencedMetaBlockHashes, s.referencedMetaBlockHashes)

	return referencedMetaBlockHashes
}

// GetReferencedMetaBlocks returns the referenced meta blocks
func (s *miniBlocksSelectionSession) GetReferencedMetaBlocks() []data.HeaderHandler {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.referencedMetaBlocks) == 0 {
		return nil
	}

	referencedMetaBlocks := make([]data.HeaderHandler, len(s.referencedMetaBlocks))
	copy(referencedMetaBlocks, s.referencedMetaBlocks)

	return referencedMetaBlocks
}

// GetLastMetaBlock returns the last meta block
func (s *miniBlocksSelectionSession) GetLastMetaBlock() data.HeaderHandler {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.lastMetaBlock
}

// GetGasProvided returns the gas provided for the mini blocks
func (s *miniBlocksSelectionSession) GetGasProvided() uint64 {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.gasProvided
}

// GetNumTxsAdded returns the number of transactions added to the mini blocks
func (s *miniBlocksSelectionSession) GetNumTxsAdded() uint32 {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.numTxsAdded
}

// AddMiniBlocksAndHashes adds a slice of mini blocks and their hashes to the session
func (s *miniBlocksSelectionSession) AddMiniBlocksAndHashes(miniBlocksAndHashes []block.MiniblockAndHash) error {
	var err error
	miniBlocks := make([]*block.MiniBlock, 0, len(miniBlocksAndHashes))
	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, 0, len(miniBlocksAndHashes))
	miniBlockHashes := make([][]byte, 0, len(miniBlocksAndHashes))
	numTxsAdded := 0

	for _, miniBlockInfo := range miniBlocksAndHashes {
		txCount := len(miniBlockInfo.Miniblock.GetTxHashes())

		mbHeader := &block.MiniBlockHeader{
			Hash:            miniBlockInfo.Hash,
			SenderShardID:   miniBlockInfo.Miniblock.SenderShardID,
			ReceiverShardID: miniBlockInfo.Miniblock.ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            miniBlockInfo.Miniblock.GetType(),
		}
		err = setProcessingTypeAndConstructionStateForProposalMb(mbHeader)
		if err != nil {
			return err
		}

		miniBlocks = append(miniBlocks, miniBlockInfo.Miniblock)
		miniBlockHeaderHandlers = append(miniBlockHeaderHandlers, mbHeader)
		miniBlockHashes = append(miniBlockHashes, miniBlockInfo.Hash)
		numTxsAdded += txCount
	}

	s.mut.Lock()
	defer s.mut.Unlock()
	s.miniBlocks = append(s.miniBlocks, miniBlocks...)
	s.miniBlockHeaderHandlers = append(s.miniBlockHeaderHandlers, miniBlockHeaderHandlers...)
	s.miniBlockHashes = append(s.miniBlockHashes, miniBlockHashes...)
	s.numTxsAdded += uint32(numTxsAdded)
	// TODO: take care of the gas management

	return nil
}

// CreateAndAddMiniBlockFromTransactions creates a mini block from the provided transaction hashes
func (s *miniBlocksSelectionSession) CreateAndAddMiniBlockFromTransactions(txHashes [][]byte) error {
	if len(txHashes) == 0 {
		return nil
	}

	// TODO: add estimated gas management

	// no need to create multiple miniblocks from the shard to itself or to other shards before processing
	// the transactions, so create a single miniBlock
	outgoingMiniBlock := &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: s.shardID,
		SenderShardID:   s.shardID,
		Type:            block.TxBlock,
		Reserved:        nil,
	}

	outgoingMiniBlockHash, err := core.CalculateHash(s.marshaller, s.hasher, outgoingMiniBlock)
	if err != nil {
		return err
	}

	return s.AddMiniBlocksAndHashes([]block.MiniblockAndHash{{
		Miniblock: outgoingMiniBlock,
		Hash:      outgoingMiniBlockHash,
	}})
}

func setProcessingTypeAndConstructionStateForProposalMb(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
) error {
	err := miniBlockHeaderHandler.SetProcessingType(int32(block.Normal))
	if err != nil {
		return err
	}

	err = miniBlockHeaderHandler.SetConstructionState(int32(block.Proposed))
	if err != nil {
		return err
	}
	return nil
}

// IsInterfaceNil checks if the interface is nil
func (s *miniBlocksSelectionSession) IsInterfaceNil() bool {
	return s == nil
}
