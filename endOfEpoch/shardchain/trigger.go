package shardchain

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
)

// ArgsNewShardEndOfEpochTrigger defines the arguments needed for new end of epoch trigger
type ArgsNewShardEndOfEpochTrigger struct {
}

type trigger struct {
}

// NewEndOfEpochTrigger creates a trigger to signal end of epoch
func NewEndOfEpochTrigger(args *ArgsNewShardEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewShardEndOfEpochTrigger
	}

	return &trigger{}, nil
}

// IsEndOfEpoch returns true if conditions are fullfilled for end of epoch
func (t *trigger) IsEndOfEpoch() bool {
	return false
}

// Epoch returns the current epoch number
func (t *trigger) Epoch() uint32 {
	return 0
}

// ForceEndOfEpoch sets the conditions for end of epoch to true in case of edge cases
func (t *trigger) ForceEndOfEpoch(round int64) error {
	return nil
}

// ReceivedHeader saves the header into pool to verify if end-of-epoch conditions are fullfilled
func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
}

// Update updates the end-of-epoch trigger
func (t *trigger) Update(round int64) {
}

// Processed signals that end-of-epoch has been processed
func (t *trigger) Processed() {
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
