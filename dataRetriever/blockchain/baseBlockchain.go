package blockchain

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
)

type baseBlockChain struct {
	mut                    sync.RWMutex
	appStatusHandler       core.AppStatusHandler
	genesisHeader          data.HeaderHandler
	genesisHeaderHash      []byte
	currentBlockHeader     data.HeaderHandler
	currentBlockHeaderHash []byte
	finalCoordinates       *finalCoordinates
}

// GetGenesisHeader returns the genesis block header pointer
func (bbc *baseBlockChain) GetGenesisHeader() data.HeaderHandler {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	if check.IfNil(bbc.genesisHeader) {
		return nil
	}

	return bbc.genesisHeader.ShallowClone()
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bbc *baseBlockChain) GetGenesisHeaderHash() []byte {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	return bbc.genesisHeaderHash
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bbc *baseBlockChain) SetGenesisHeaderHash(hash []byte) {
	bbc.mut.Lock()
	bbc.genesisHeaderHash = hash
	bbc.mut.Unlock()
}

// GetCurrentBlockHeader returns current block header pointer
func (bbc *baseBlockChain) GetCurrentBlockHeader() data.HeaderHandler {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	if check.IfNil(bbc.currentBlockHeader) {
		return nil
	}

	return bbc.currentBlockHeader.ShallowClone()
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bbc *baseBlockChain) GetCurrentBlockHeaderHash() []byte {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	return bbc.currentBlockHeaderHash
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bbc *baseBlockChain) SetCurrentBlockHeaderHash(hash []byte) {
	bbc.mut.Lock()
	bbc.currentBlockHeaderHash = hash
	bbc.mut.Unlock()
}

func (bbc *baseBlockChain) SetHighestFinalBlockCoordinates(nonce uint64, headerHash []byte, rootHash []byte) {
	bbc.mut.Lock()

	bbc.finalCoordinates.blockNonce = nonce
	bbc.finalCoordinates.blockHash = headerHash
	bbc.finalCoordinates.blockRootHash = rootHash

	bbc.mut.Unlock()
}

func (bbc *baseBlockChain) GetHighestFinalBlockCoordinates() (nonce uint64, hash []byte, rootHash []byte) {
	bbc.mut.RLock()

	nonce = bbc.finalCoordinates.blockNonce
	hash = bbc.finalCoordinates.blockHash
	rootHash = bbc.finalCoordinates.blockRootHash

	bbc.mut.RUnlock()

	return
}
