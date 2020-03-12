package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

type ValidatorInfoProcessorStub struct {
}

func (vip *ValidatorInfoProcessorStub) TryProcessMetaBlock(*block.MetaBlock, []byte) error {
	return nil
}

func (vip *ValidatorInfoProcessorStub) IsInterfaceNil() bool {
	return vip == nil
}
