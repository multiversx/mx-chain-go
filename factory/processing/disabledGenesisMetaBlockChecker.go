package processing

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

type disabledGenesisMetaBlockChecker struct {
}

// NewDisabledGenesisMetaBlockChecker creates a disabled meta block genesis checker
func NewDisabledGenesisMetaBlockChecker() *disabledGenesisMetaBlockChecker {
	return &disabledGenesisMetaBlockChecker{}
}

// CheckGenesisMetaBlock does nothing
func (gmbc *disabledGenesisMetaBlockChecker) CheckGenesisMetaBlock(_ map[uint32]data.HeaderHandler, _ []byte) error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gmbc *disabledGenesisMetaBlockChecker) IsInterfaceNil() bool {
	return gmbc == nil
}
