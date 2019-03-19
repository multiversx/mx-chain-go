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
// The BlockChain also holds pointers to the Genesis block header, the current block
// the height of the local chain and the perceived height of the chain in the network.
type BlockChain struct {
	StorageService
	GenesisBlockHeader     *block.Header  // Genesys Block Header pointer
	GenesisHeaderHash      []byte         // Genesis Block Header hash
	CurrentBlockHeader     *block.Header  // Current Block pointer
	CurrentBlockHeaderHash []byte         // Current Block Header hash
	CurrentTxBlockBody     block.Body     // Current Tx Block Body pointer
	LocalHeight            int64          // Height of the local chain
	NetworkHeight          int64          // Percieved height of the network chain
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
		GenesisBlockHeader: nil,
		CurrentBlockHeader: nil,
		LocalHeight:        -1,
		NetworkHeight:      -1,
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

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChain) IsBadBlock(blockHash []byte) bool {
	return bc.badBlocks.Has(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChain) PutBadBlock(blockHash []byte) {
	bc.badBlocks.Put(blockHash, struct{}{})
}
