package blockchain

import (
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
	StorageService
	genesisBlock           *block.Header
	genesisHeaderHash      []byte
	currentBlockHeader     *block.Header
	currentBlockHeaderHash []byte
	currentTxBlockBody     block.Body
	localHeight            int64
	networkHeight          int64
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

	data := &BlockChain{
		genesisBlock:       nil,
		currentBlockHeader: nil,
		localHeight:        -1,
		networkHeight:      -1,
		badBlocks:          badBlocksCache,
		StorageService: &ChainStorer{
			chain: map[UnitType]storage.Storer{
				TransactionUnit: txUnit,
				MiniBlockUnit:   miniBlockUnit,
				PeerChangesUnit: peerChangesBlockUnit,
				BlockHeaderUnit: headerUnit,
			},
		},
	}

	return data, nil
}

// GenesisBlock returns the genesis block header pointer
func (bc *BlockChain) GenesisBlock() *block.Header {
	return bc.genesisBlock
}

// SetGenesisBlock sets the genesis block header pointer
func (bc *BlockChain) SetGenesisBlock(genesisBlock *block.Header) {
	bc.genesisBlock = genesisBlock
}

// GenesisHeaderHash returns the genesis block header hash
func (bc *BlockChain) GenesisHeaderHash() []byte {
	return bc.genesisHeaderHash
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChain) SetGenesisHeaderHash(hash []byte) {
	bc.genesisHeaderHash = hash
}

// CurrentBlockHeader returns current block header pointer
func (bc *BlockChain) CurrentBlockHeader() *block.Header {
	return bc.currentBlockHeader
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *BlockChain) SetCurrentBlockHeader(header *block.Header) {
	bc.currentBlockHeader = header
}

// CurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) CurrentBlockHeaderHash() []byte {
	return bc.currentBlockHeaderHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChain) SetCurrentBlockHeaderHash(hash []byte) {
	bc.currentBlockHeaderHash = hash
}

// CurrentTxBlockBody returns the tx block body pointer
func (bc *BlockChain) CurrentTxBlockBody() block.Body {
	return bc.currentTxBlockBody
}

// SetCurrentTxBlockBody sets the tx block body pointer
func (bc *BlockChain) SetCurrentTxBlockBody(body block.Body) {
	bc.currentTxBlockBody = body
}

// LocalHeight returns the height of the local chain
func (bc *BlockChain) LocalHeight() int64 {
	return bc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (bc *BlockChain) SetLocalHeight(height int64) {
	bc.localHeight = height
}

// NetworkHeight sets the percieved height of the network chain
func (bc *BlockChain) NetworkHeight() int64 {
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
