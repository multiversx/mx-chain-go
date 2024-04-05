package process

import (
	"github.com/multiversx/mx-chain-go/errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignGenesisMetaBlockChecker struct {
}

// NewSovereignGenesisMetaBlockChecker creates a sovereign meta block genesis checker
func NewSovereignGenesisMetaBlockChecker() *sovereignGenesisMetaBlockChecker {
	return &sovereignGenesisMetaBlockChecker{}
}

// SetValidatorRootHashOnGenesisMetaBlock checks that the genesis meta block does not exist
func (gmbc *sovereignGenesisMetaBlockChecker) SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, _ []byte) error {
	if !check.IfNil(genesisMetaBlock) {
		return errors.ErrGenesisMetaBlockOnSovereign
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *sovereignGenesisMetaBlockChecker) IsInterfaceNil() bool {
	return gmbc == nil
}
