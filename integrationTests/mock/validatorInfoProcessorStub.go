package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

// ValidatorInfoProcessorStub -
type ValidatorInfoProcessorStub struct {
}

// ProcessMetaBlock -
func (vip *ValidatorInfoProcessorStub) ProcessMetaBlock(*block.MetaBlock, []byte) ([][]byte, error) {
	return nil, nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoProcessorStub) IsInterfaceNil() bool {
	return vip == nil
}
