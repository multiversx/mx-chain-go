package processing

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

type genesisMetaBlockChecker struct {
}

// NewGenesisMetaBlockChecker creates a new meta genesis block checker
func NewGenesisMetaBlockChecker() *genesisMetaBlockChecker {
	return &genesisMetaBlockChecker{}
}

// CheckGenesisMetaBlock checks if the genesis blocks contain the meta block and sets the corresponding validator root hash
func (gmbc *genesisMetaBlockChecker) CheckGenesisMetaBlock(
	genesisBlocks map[uint32]data.HeaderHandler,
	validatorStatsRootHash []byte,
) error {
	genesisBlock, ok := genesisBlocks[core.MetachainShardId]
	if !ok {
		return errors.New("genesis meta block does not exist")
	}

	genesisMetaBlock, ok := genesisBlock.(data.MetaHeaderHandler)
	if !ok {
		return errors.New("genesis meta block invalid")
	}

	return genesisMetaBlock.SetValidatorStatsRootHash(validatorStatsRootHash)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *genesisMetaBlockChecker) IsInterfaceNil() bool {
	return gmbc == nil
}
