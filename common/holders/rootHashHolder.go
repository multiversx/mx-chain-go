package holders

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
)

type rootHashHolder struct {
	rootHash []byte
	epoch    core.OptionalUint32
}

// NewRootHashHolder creates a rootHashHolder
func NewRootHashHolder(rootHash []byte, epoch core.OptionalUint32) *rootHashHolder {
	return &rootHashHolder{
		rootHash: rootHash,
		epoch:    epoch,
	}
}

// NewRootHashHolderAsEmpty creates an empty rootHashHolder
func NewRootHashHolderAsEmpty() *rootHashHolder {
	return &rootHashHolder{
		rootHash: nil,
		epoch:    core.OptionalUint32{},
	}
}

// GetRootHash returns the contained rootHash
func (holder *rootHashHolder) GetRootHash() []byte {
	return holder.rootHash
}

// GetEpoch returns the epoch of the contained rootHash
func (holder *rootHashHolder) GetEpoch() core.OptionalUint32 {
	return holder.epoch
}

// String returns rootHashesHolder as a string
func (holder *rootHashHolder) String() string {
	return fmt.Sprintf("root hash %s, epoch %v, has value %v", holder.rootHash, holder.epoch.Value, holder.epoch.HasValue)
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *rootHashHolder) IsInterfaceNil() bool {
	return holder == nil
}
