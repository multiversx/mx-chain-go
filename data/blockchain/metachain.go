package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// MetaChain holds the block information for the beacon chain
//
// The MetaChain through it's Storage units map manages the storage,
// retrieval search of blocks (body), transactions, block headers,
// bad blocks.
//
// The MetaChain also holds pointers to the Genesis block, the current block
// the height of the local chain and the perceived height of the chain in the network.
type MetaChain struct {
	StorageService
	genesisBlock     *block.MetaBlock // Genesys Block pointer
	genesisBlockHash []byte           // Genesis Block hash
	currentBlock     *block.MetaBlock // Current Block pointer
	currentBlockHash []byte           // Current Block hash
	localHeight      int64            // Height of the local chain
	networkHeight    int64            // Percieved height of the network chain
	badBlocks        storage.Cacher   // Bad blocks cache
}

// NewMetachain will initialize a new metachain instance
func NewMetaChain(
	firstBlock *block.MetaBlock,
	firstBlockHash []byte,
	badBlocksCache storage.Cacher,
	metaBlockUnit storage.Storer,
	shardDataUnit storage.Storer,
	peerDataUnit storage.Storer) (*MetaChain, error) {

	if firstBlock == nil {
		return nil, ErrMetaGenesisBlockNil
	}

	if firstBlockHash == nil {
		return nil, ErrMetaGenesisBlockHashNil
	}

	if badBlocksCache == nil {
		return nil, ErrBadBlocksCacheNil
	}

	if metaBlockUnit == nil {
		return nil, ErrMetaBlockUnitNil
	}

	if shardDataUnit == nil {
		return nil, ErrShardDataUnitNil
	}

	if peerDataUnit == nil {
		return nil, ErrPeerDataUnitNil
	}

	return &MetaChain{
		badBlocks: badBlocksCache,
		StorageService: &ChainStorer{
			chain: map[UnitType]storage.Storer{
				MetaBlockUnit:     metaBlockUnit,
				MetaShardDataUnit: shardDataUnit,
				MetaPeerDataUnit:  peerDataUnit,
			},
		},
	}, nil
}

// GenesisBlock returns the genesis block header pointer
func (mc *MetaChain) GenesisBlock() *block.MetaBlock {
	return mc.genesisBlock
}

// GenesisHeaderHash returns the genesis block header hash
func (mc *MetaChain) GenesisHeaderHash() []byte {
	return mc.genesisBlockHash
}

// CurrentBlockHeader returns current block header pointer
func (mc *MetaChain) CurrentBlockHeader() *block.MetaBlock {
	return mc.currentBlock
}

// SetCurrentBlockHeader sets current block header pointer
func (mc *MetaChain) SetCurrentBlockHeader(header *block.MetaBlock) {
	mc.currentBlock = header
}

// CurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) CurrentBlockHeaderHash() []byte {
	return mc.currentBlockHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) SetCurrentBlockHeaderHash(hash []byte) {
	mc.currentBlockHash = hash
}

// CurrentBlockBody returns the block body pointer
func (mc *MetaChain) CurrentBlockBody() *block.MetaBlock {
	return nil
}

// SetCurrentBlockBody sets the block body pointer
func (mc *MetaChain) SetCurrentBlockBody(body *block.MetaBlock) {
	// not needed to be implemented in metachain.
}

// LocalHeight returns the height of the local chain
func (mc *MetaChain) LocalHeight() int64 {
	return mc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (mc *MetaChain) SetLocalHeight(height int64) {
	mc.localHeight = height
}

// NetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) NetworkHeight() int64 {
	return mc.localHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) SetNetworkHeight(height int64) {
	mc.localHeight = height
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (mc *MetaChain) IsBadBlock(blockHash []byte) bool {
	return mc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (mc *MetaChain) PutBadBlock(blockHash []byte) {
	mc.badBlocks.Put(blockHash, struct{}{})
}
