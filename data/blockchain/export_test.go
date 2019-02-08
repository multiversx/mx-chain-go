package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func NewChainStorer() *ChainStorer {
	return &ChainStorer{
		chain: make(map[UnitType]storage.Storer),
	}
}
func (bc *ChainStorer) AddStorer(key UnitType, s storage.Storer) {
	bc.lock.Lock()
	bc.chain[key] = s
	bc.lock.Unlock()
}