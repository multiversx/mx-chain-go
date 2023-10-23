package dblookupext

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

func (hr *historyRepository) computeMiniblockHash(miniblock *block.MiniBlock) []byte {
	hash, err := core.CalculateHash(hr.marshalizer, hr.hasher, miniblock)
	log.LogIfError(err)

	return hash
}
