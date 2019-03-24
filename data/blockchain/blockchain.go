package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// BlockChain holds the block information for the current shard.
//
// The BlockChain through it's Storage units map manages the storage,
// retrieval search of blocks (body), transactions, block headers,
// bad blocks.
//
// The BlockChain also holds pointers to the Genesis block header, the current block
// the height of the local chain and the perceived height of the chain in the network.
type BlockChain struct {
	data.StorageService
	GenesisHeader          *block.Header  // Genesis Block Header pointer
	genesisHeaderHash      []byte         // Genesis Block Header hash
	CurrentBlockHeader     *block.Header  // Current Block Header pointer
	currentBlockHeaderHash []byte         // Current Block Header hash
	CurrentBlockBody       block.Body     // Current Block Body pointer
	localHeight            int64          // Height of the local chain
	networkHeight          int64          // Percieved height of the network chain
	badBlocks              storage.Cacher // Bad blocks cache
}

// NewBlockChain returns an initialized blockchain
// It uses a config file to setup it's supported storage units map
func NewBlockChain(
	badBlocksCache storage.Cacher,
	txUnit storage.Storer,
	miniBlockUnit storage.Storer,
	peerChangesBlockUnit storage.Storer,
	headerUnit storage.Storer,
	metachainHeadersUnit storage.Storer,
) (*BlockChain, error) {

	if badBlocksCache == nil {
		return nil, ErrBadBlocksCacheNil
	}
	if txUnit == nil {
		return nil, ErrTxUnitNil
	}
	if miniBlockUnit == nil {
		return nil, ErrMiniBlockUnitNil
	}
	if peerChangesBlockUnit == nil {
		return nil, ErrPeerBlockUnitNil
	}
	if headerUnit == nil {
		return nil, ErrHeaderUnitNil
	}
	if metachainHeadersUnit == nil {
		return nil, ErrMetachainHeaderUnitNil
	}

	blockChain := &BlockChain{
		GenesisHeader:      nil,
		CurrentBlockHeader: nil,
		localHeight:        -1,
		networkHeight:      -1,
		badBlocks:          badBlocksCache,
		StorageService: &ChainStorer{
			chain: map[data.UnitType]storage.Storer{
				data.TransactionUnit: txUnit,
				data.MiniBlockUnit:   miniBlockUnit,
				data.PeerChangesUnit: peerChangesBlockUnit,
				data.BlockHeaderUnit: headerUnit,
				data.MetaBlockUnit:   metachainHeadersUnit,
			},
		},
	}

	return blockChain, nil
}

// GetGenesisHeader returns the genesis block header pointer
func (bc *BlockChain) GetGenesisHeader() data.HeaderHandler {
	if bc.GenesisHeader == nil {
		return nil
	}
	return bc.GenesisHeader
}

// SetGenesisHeader sets the genesis block header pointer
func (bc *BlockChain) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if genesisBlock == nil {
		bc.GenesisHeader = nil
		return nil
	}

	gb, ok := genesisBlock.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.GenesisHeader = gb
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bc *BlockChain) GetGenesisHeaderHash() []byte {
	return bc.genesisHeaderHash
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChain) SetGenesisHeaderHash(hash []byte) {
	bc.genesisHeaderHash = hash
}

// GetCurrentBlockHeader returns current block header pointer
func (bc *BlockChain) GetCurrentBlockHeader() data.HeaderHandler {
	if bc.CurrentBlockHeader == nil {
		return nil
	}
	return bc.CurrentBlockHeader
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *BlockChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if header == nil {
		bc.CurrentBlockHeader = nil
		return nil
	}

	h, ok := header.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.CurrentBlockHeader = h
	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) GetCurrentBlockHeaderHash() []byte {
	return bc.currentBlockHeaderHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) SetCurrentBlockHeaderHash(hash []byte) {
	bc.currentBlockHeaderHash = hash
}

// GetCurrentBlockBody returns the tx block body pointer
func (bc *BlockChain) GetCurrentBlockBody() data.BodyHandler {
	if bc.CurrentBlockBody == nil {
		return nil
	}
	return bc.CurrentBlockBody
}

// SetCurrentBlockBody sets the tx block body pointer
func (bc *BlockChain) SetCurrentBlockBody(body data.BodyHandler) error {
	if body == nil {
		bc.CurrentBlockBody = nil
		return nil
	}

	blockBody, ok := body.(block.Body)
	if !ok {
		return data.ErrInvalidBodyType
	}
	bc.CurrentBlockBody = blockBody
	return nil
}

// GetLocalHeight returns the height of the local chain
func (bc *BlockChain) GetLocalHeight() int64 {
	return bc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (bc *BlockChain) SetLocalHeight(height int64) {
	bc.localHeight = height
}

// GetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChain) GetNetworkHeight() int64 {
	return bc.localHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChain) SetNetworkHeight(height int64) {
	bc.localHeight = height
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChain) PutBadBlock(blockHash []byte) {
	bc.badBlocks.Put(blockHash, struct{}{})
}
