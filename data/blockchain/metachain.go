package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
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
	GenesisBlock     *block.MetaBlock // Genesys Block pointer
	genesisBlockHash []byte           // Genesis Block hash
	CurrentBlock     *block.MetaBlock // Current Block pointer
	currentBlockHash []byte           // Current Block hash
	localHeight      int64            // Height of the local chain
	networkHeight    int64            // Percieved height of the network chain
	badBlocks        storage.Cacher   // Bad blocks cache
}

// NewMetachain will initialize a new metachain instance
func NewMetaChain(
	badBlocksCache storage.Cacher,
	metaBlockUnit storage.Storer,
	shardDataUnit storage.Storer,
	peerDataUnit storage.Storer) (*MetaChain, error) {
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

// GetGenesisBlock returns the genesis block header pointer
func (mc *MetaChain) GetGenesisHeader() data.HeaderHandler {
	if mc.GenesisBlock == nil {
		return nil
	}
	return mc.GenesisBlock
}

// SetGenesisBlock returns the genesis block header pointer
func (mc *MetaChain) SetGenesisHeader(header data.HeaderHandler) error {
	genBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}
	mc.GenesisBlock = genBlock
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (mc *MetaChain) GetGenesisHeaderHash() []byte {
	return mc.genesisBlockHash
}

// SetGenesisHeaderHash returns the genesis block header hash
func (mc *MetaChain) SetGenesisHeaderHash(headerHash []byte) {
	mc.genesisBlockHash = headerHash
}

// GetCurrentBlockHeader returns current block header pointer
func (mc *MetaChain) GetCurrentBlockHeader() data.HeaderHandler {
	if mc.CurrentBlock == nil {
		return nil
	}
	return mc.CurrentBlock
}

// SetCurrentBlockHeader sets current block header pointer
func (mc *MetaChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	currHead, ok := header.(*block.MetaBlock)
	if !ok {
		return ErrWrongTypeInSet
	}
	mc.CurrentBlock = currHead
	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) GetCurrentBlockHeaderHash() []byte {
	return mc.currentBlockHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (mc *MetaChain) SetCurrentBlockHeaderHash(hash []byte) {
	mc.currentBlockHash = hash
}

// GetCurrentBlockBody returns the block body pointer
func (mc *MetaChain) GetCurrentBlockBody() data.BodyHandler {
	return nil
}

// SetCurrentBlockBody sets the block body pointer
func (mc *MetaChain) SetCurrentBlockBody(body data.BodyHandler) error {
	// not needed to be implemented in metachain.
	return nil
}

// GetLocalHeight returns the height of the local chain
func (mc *MetaChain) GetLocalHeight() int64 {
	return mc.localHeight
}

// SetLocalHeight sets the height of the local chain
func (mc *MetaChain) SetLocalHeight(height int64) {
	mc.localHeight = height
}

// GetNetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) GetNetworkHeight() int64 {
	return mc.networkHeight
}

// SetNetworkHeight sets the percieved height of the network chain
func (mc *MetaChain) SetNetworkHeight(height int64) {
	mc.networkHeight = height
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (mc *MetaChain) IsBadBlock(blockHash []byte) bool {
	return mc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (mc *MetaChain) PutBadBlock(blockHash []byte) {
	mc.badBlocks.Put(blockHash, struct{}{})
}
