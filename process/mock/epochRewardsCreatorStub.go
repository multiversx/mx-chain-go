package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type EpochRewardsCreatorStub struct {
}

func (e *EpochRewardsCreatorStub) CreateBlockStarted() {
}

func (e *EpochRewardsCreatorStub) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) (data.BodyHandler, error) {
	return make(block.Body, 0), nil
}

func (e *EpochRewardsCreatorStub) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) error {
	return nil
}

func (e *EpochRewardsCreatorStub) CreateMarshalizedData() map[string][][]byte {
	return nil
}

func (e *EpochRewardsCreatorStub) SaveTxBlockToStorage(body block.Body) error {
	return nil
}

func (e *EpochRewardsCreatorStub) DeleteTxsFromStorage(body block.Body) error {
	return nil
}

func (e *EpochRewardsCreatorStub) IsInterfaceNil() bool {
	return e == nil
}
