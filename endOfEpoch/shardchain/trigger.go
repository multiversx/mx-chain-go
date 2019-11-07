package shardchain

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
)

type ArgsNewShardEndOfEpochTrigger struct {
}

type trigger struct {
}

func NewEndOfEpochTrigger(args *ArgsNewShardEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewShardEndOfEpochTrigger
	}

	return &trigger{}, nil
}

func (t *trigger) IsEndOfEpoch() bool {

	return false
}

func (t *trigger) Epoch() uint32 {
	return 0
}

func (t *trigger) ForceEndOfEpoch() {

}

func (t *trigger) ReceivedHeader(header data.HeaderHandler) {

}

func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
