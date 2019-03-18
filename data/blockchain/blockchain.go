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
// The BlockChain also holds pointers to the Genesis block, the current block
// the height of the local chain and the perceived height of the chain in the network.
type BlockChain struct {
	data.StorageService
	GenesisHeader          *block.Header
	GenesisHeaderHash      []byte
	CurrentBlockHeader     *block.Header
	CurrentBlockHeaderHash []byte
	CurrentBlockBody       block.Body
	LocalHeight            int64
	NetworkHeight          int64
	badBlocks              storage.Cacher // Bad blocks cache
}

// NewBlockChain returns an initialized blockchain
// It uses a config file to setup it's supported storage units map
func NewBlockChain(
	badBlocksCache storage.Cacher,
	txUnit storage.Storer,
	miniBlockUnit storage.Storer,
	peerChangesBlockUnit storage.Storer,
	headerUnit storage.Storer) (*BlockChain, error) {

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

	blockChain := &BlockChain{
		GenesisHeader:      nil,
		CurrentBlockHeader: nil,
		LocalHeight:        -1,
		NetworkHeight:      -1,
		badBlocks:          badBlocksCache,
		StorageService: &ChainStorer{
			chain: map[data.UnitType]storage.Storer{
				data.TransactionUnit: txUnit,
				data.MiniBlockUnit:   miniBlockUnit,
				data.PeerChangesUnit: peerChangesBlockUnit,
				data.BlockHeaderUnit: headerUnit,
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
	gb, ok := genesisBlock.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.GenesisHeader = gb
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bc *BlockChain) GetGenesisHeaderHash() []byte {
	return bc.GenesisHeaderHash
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChain) SetGenesisHeaderHash(hash []byte) {
	bc.GenesisHeaderHash = hash
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
	h, ok := header.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.CurrentBlockHeader = h
	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) GetCurrentBlockHeaderHash() []byte {
	return bc.CurrentBlockHeaderHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) SetCurrentBlockHeaderHash(hash []byte) {
	bc.CurrentBlockHeaderHash = hash
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
	blockBody, ok := body.(block.Body)
	if !ok {
		return data.ErrInvalidBodyType
	}
	bc.CurrentBlockBody = blockBody
	return nil
}

// GetLocalHeight returns the height of the local chain
func (bc *BlockChain) GetLocalHeight() int64 {
	return bc.LocalHeight
}

// SetLocalHeight sets the height of the local chain
func (bc *BlockChain) SetLocalHeight(height int64) {
	bc.LocalHeight = height
}

// GetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChain) GetNetworkHeight() int64 {
	return bc.LocalHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChain) SetNetworkHeight(height int64) {
	bc.LocalHeight = height
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChain) PutBadBlock(blockHash []byte) {
	bc.badBlocks.Put(blockHash, struct{}{})
}
