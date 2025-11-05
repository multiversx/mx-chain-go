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
	miniBlockHeaderHandlers     []data.MiniBlockHeaderHandler
	miniBlocks                  block.MiniBlockSlice
	miniBlockHashes             [][]byte
	miniBlockHashesUnique       map[string]struct{}
	referenceHeaderHashesUnique map[string]struct{}
	referencedHeaderHashes      [][]byte
	referencedHeader            []data.HeaderHandler
	lastHeader                  data.HeaderHandler
	numTxsAdded                 uint32
	mut                         sync.RWMutex
}

type miniBlocksSelectionSession struct {
	*miniBlocksSelectionConfig
	*miniBlocksSelectionResult
}

const defaultCapacity = 10

// NewMiniBlocksSelectionSession creates a new instance of miniBlocksSelectionSession
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
			miniBlockHeaderHandlers:     make([]data.MiniBlockHeaderHandler, 0, defaultCapacity),
			miniBlocks:                  make(block.MiniBlockSlice, 0, defaultCapacity),
			miniBlockHashes:             make([][]byte, 0, defaultCapacity),
			miniBlockHashesUnique:       make(map[string]struct{}),
			referenceHeaderHashesUnique: make(map[string]struct{}),
			referencedHeaderHashes:      make([][]byte, 0, defaultCapacity),
			referencedHeader:            make([]data.HeaderHandler, 0, defaultCapacity),
			lastHeader:                  nil,
			numTxsAdded:                 0,
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
	s.miniBlockHashesUnique = make(map[string]struct{})
	s.referenceHeaderHashesUnique = make(map[string]struct{})
	s.referencedHeaderHashes = make([][]byte, 0, defaultCapacity)
	s.referencedHeader = make([]data.HeaderHandler, 0, defaultCapacity)
	s.lastHeader = nil
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

// AddReferencedHeader adds a header and its hash to the session
func (s *miniBlocksSelectionSession) AddReferencedHeader(header data.HeaderHandler, headerHash []byte) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if check.IfNil(header) || len(headerHash) == 0 {
		return
	}

	_, ok := s.referenceHeaderHashesUnique[string(headerHash)]
	if !ok {
		s.referenceHeaderHashesUnique[string(headerHash)] = struct{}{}
		s.referencedHeader = append(s.referencedHeader, header)
		s.referencedHeaderHashes = append(s.referencedHeaderHashes, headerHash)
		s.lastHeader = header
	}
}

// GetReferencedHeaderHashes returns the hashes of the referenced headers
func (s *miniBlocksSelectionSession) GetReferencedHeaderHashes() [][]byte {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.referencedHeaderHashes) == 0 {
		return nil
	}

	referencedHeaderHashes := make([][]byte, len(s.referencedHeaderHashes))
	copy(referencedHeaderHashes, s.referencedHeaderHashes)

	return referencedHeaderHashes
}

// GetReferencedHeaders returns the referenced headers
func (s *miniBlocksSelectionSession) GetReferencedHeaders() []data.HeaderHandler {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if len(s.referencedHeader) == 0 {
		return nil
	}

	referencedHeaders := make([]data.HeaderHandler, len(s.referencedHeader))
	copy(referencedHeaders, s.referencedHeader)

	return referencedHeaders
}

// GetLastHeader returns the last header
func (s *miniBlocksSelectionSession) GetLastHeader() data.HeaderHandler {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.lastHeader
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

	s.mut.Lock()
	defer s.mut.Unlock()

	for _, miniBlockInfo := range miniBlocksAndHashes {
		_, ok := s.miniBlockHashesUnique[string(miniBlockInfo.Hash)]
		if ok {
			continue
		}

		s.miniBlockHashesUnique[string(miniBlockInfo.Hash)] = struct{}{}

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

	s.miniBlocks = append(s.miniBlocks, miniBlocks...)
	s.miniBlockHeaderHandlers = append(s.miniBlockHeaderHandlers, miniBlockHeaderHandlers...)
	s.miniBlockHashes = append(s.miniBlockHashes, miniBlockHashes...)
	s.numTxsAdded += uint32(numTxsAdded)

	return nil
}

// CreateAndAddMiniBlockFromTransactions creates a mini block from the provided transaction hashes
func (s *miniBlocksSelectionSession) CreateAndAddMiniBlockFromTransactions(txHashes [][]byte) error {
	if len(txHashes) == 0 {
		return nil
	}

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
