package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// blockChain holds the block information for the current shard.
//
// The blockChain through it's Storage units map manages the storage,
// retrieval search of blocks (body), transactions, block headers,
// bad blocks.
//
// The blockChain also holds pointers to the Genesis block, the current block
// the height of the local chain and the perceived height of the chain in the network.
type blockChain struct {
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
	headerUnit storage.Storer) (*blockChain, error) {

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

	data := &blockChain{
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
func (bc *blockChain) GenesisBlock() *block.Header {
	return bc.genesisBlock
}

// SetGenesisBlock sets the genesis block header pointer
func (bc *blockChain) SetGenesisBlock(genesisBlock *block.Header) {
	bc.genesisBlock = genesisBlock
}

// GenesisHeaderHash returns the genesis block header hash
func (bc *blockChain) GenesisHeaderHash() []byte {
	return bc.genesisHeaderHash
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *blockChain) SetGenesisHeaderHash(hash []byte) {
	bc.genesisHeaderHash = hash
}

// CurrentBlockHeader returns current block header pointer
func (bc *blockChain) CurrentBlockHeader() *block.Header {
	return bc.currentBlockHeader
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *blockChain) SetCurrentBlockHeader(header *block.Header) {
	bc.currentBlockHeader = header
}

// CurrentBlockHeaderHash returns the current block header hash
func (bc *blockChain) CurrentBlockHeaderHash() []byte {
	return bc.currentBlockHeaderHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *blockChain) SetCurrentBlockHeaderHash(hash []byte) {
	bc.currentBlockHeaderHash = hash
}

// CurrentTxBlockBody returns the tx block body pointer
func (bc *blockChain) CurrentTxBlockBody() block.Body {
	return bc.currentTxBlockBody
}

// SetCurrentTxBlockBody sets the tx block body pointer
func (bc *blockChain) SetCurrentTxBlockBody(body block.Body) {
	bc.currentTxBlockBody = body
}

// LocalHeight returns the height of the local chain
func (bc *blockChain) LocalHeight() int64 {
	return bc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (bc *blockChain) SetLocalHeight(height int64) {
	bc.localHeight = height
}

// NetworkHeight sets the percieved height of the network chain
func (bc *blockChain) NetworkHeight() int64 {
	return bc.localHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (bc *blockChain) SetNetworkHeight(height int64) {
	bc.localHeight = height
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *blockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *blockChain) PutBadBlock(blockHash []byte) {
	bc.badBlocks.Put(blockHash, struct{}{})
}
