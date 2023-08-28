package processing

import "github.com/multiversx/mx-chain-core-go/data"

// GenesisMetaBlockChecker should handle genesis meta block checks after creation
type GenesisMetaBlockChecker interface {
	SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, validatorStatsRootHash []byte) error
	IsInterfaceNil() bool
}
