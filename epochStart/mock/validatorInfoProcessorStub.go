package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

// ValidatorInfoProcessorStub -
type ValidatorInfoProcessorStub struct {
}

// ProcessMetaBlock -
func (vip *ValidatorInfoProcessorStub) ProcessMetaBlock(*block.MetaBlock, []byte) error {
	return nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoProcessorStub) IsInterfaceNil() bool {
	return vip == nil
}
