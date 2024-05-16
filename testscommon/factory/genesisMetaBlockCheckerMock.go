package factory

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// GenesisMetaBlockCheckerMock -
type GenesisMetaBlockCheckerMock struct {
	SetValidatorRootHashOnGenesisMetaBlockCalled func(genesisMetaBlock data.HeaderHandler, validatorStatsRootHash []byte) error
}

// SetValidatorRootHashOnGenesisMetaBlock -
func (gmbc *GenesisMetaBlockCheckerMock) SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, validatorStatsRootHash []byte) error {
	if gmbc.SetValidatorRootHashOnGenesisMetaBlockCalled != nil {
		return gmbc.SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock, validatorStatsRootHash)
	}
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *GenesisMetaBlockCheckerMock) IsInterfaceNil() bool {
	return gmbc == nil
}
