package processing

import "github.com/multiversx/mx-chain-core-go/data"

// GenesisMetaBlockChecker should handle genesis meta block checks after creation
type GenesisMetaBlockChecker interface {
	CheckGenesisMetaBlock(genesisBlocks map[uint32]data.HeaderHandler, validatorStatsRootHash []byte) error
	IsInterfaceNil() bool
}
