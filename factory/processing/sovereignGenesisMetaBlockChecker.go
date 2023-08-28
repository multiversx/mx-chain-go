package processing

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/errors"
)

type disabledGenesisMetaBlockChecker struct {
}

// NewDisabledGenesisMetaBlockChecker creates a disabled meta block genesis checker
func NewDisabledGenesisMetaBlockChecker() *disabledGenesisMetaBlockChecker {
	return &disabledGenesisMetaBlockChecker{}
}

// SetValidatorRootHashOnGenesisMetaBlock checks that the genesis meta block does not exist
func (gmbc *disabledGenesisMetaBlockChecker) SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, _ []byte) error {
	if !check.IfNil(genesisMetaBlock) {
		return errors.ErrGenesisMetaBlockOnSovereign
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *disabledGenesisMetaBlockChecker) IsInterfaceNil() bool {
	return gmbc == nil
}
