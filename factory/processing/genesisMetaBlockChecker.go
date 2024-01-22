package processing

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/errors"
)

type genesisMetaBlockChecker struct {
}

// NewGenesisMetaBlockChecker creates a new meta genesis block checker
func NewGenesisMetaBlockChecker() *genesisMetaBlockChecker {
	return &genesisMetaBlockChecker{}
}

// SetValidatorRootHashOnGenesisMetaBlock checks if the genesis meta block exists and sets the corresponding validator root hash
func (gmbc *genesisMetaBlockChecker) SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, validatorStatsRootHash []byte) error {
	if check.IfNil(genesisMetaBlock) {
		return errors.ErrGenesisMetaBlockDoesNotExist
	}

	genesisMetaBlockHeader, ok := genesisMetaBlock.(data.MetaHeaderHandler)
	if !ok {
		return errors.ErrInvalidGenesisMetaBlock
	}

	return genesisMetaBlockHeader.SetValidatorStatsRootHash(validatorStatsRootHash)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *genesisMetaBlockChecker) IsInterfaceNil() bool {
	return gmbc == nil
}
