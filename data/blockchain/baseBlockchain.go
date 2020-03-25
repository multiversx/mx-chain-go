package blockchain

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type baseBlockChain struct {
	mut                    sync.RWMutex
	appStatusHandler       core.AppStatusHandler
	genesisHeader          data.HeaderHandler
	genesisHeaderHash      []byte
	currentBlockHeader     data.HeaderHandler
	currentBlockHeaderHash []byte
	currentBlockBody       data.BodyHandler
}

// SetAppStatusHandler will set the AppStatusHandler which will be used for monitoring
func (bbc *baseBlockChain) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if check.IfNil(ash) {
		return ErrNilAppStatusHandler
	}

	bbc.mut.Lock()
	bbc.appStatusHandler = ash
	bbc.mut.Unlock()
	return nil
}

// GetGenesisHeader returns the genesis block header pointer
func (bbc *baseBlockChain) GetGenesisHeader() data.HeaderHandler {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	if check.IfNil(bbc.genesisHeader) {
		return nil
	}

	return bbc.genesisHeader.Clone()
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

	return bbc.currentBlockHeader.Clone()
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

// GetCurrentBlockBody returns the tx block body pointer
func (bbc *baseBlockChain) GetCurrentBlockBody() data.BodyHandler {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()

	if check.IfNil(bbc.currentBlockBody) {
		return nil
	}

	return bbc.currentBlockBody.Clone()
}

// SetCurrentBlockBody sets the tx block body pointer
func (bbc *baseBlockChain) SetCurrentBlockBody(body data.BodyHandler) error {
	if check.IfNil(body) {
		bbc.mut.Lock()
		bbc.currentBlockBody = nil
		bbc.mut.Unlock()

		return nil
	}

	blockBody, ok := body.(*block.Body)
	if !ok {
		return data.ErrInvalidBodyType
	}

	bbc.mut.Lock()
	bbc.currentBlockBody = blockBody
	bbc.mut.Unlock()

	return nil
}
