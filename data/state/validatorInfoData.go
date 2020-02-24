package state

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/display"
)

// ValidatorInfoData is the exported data from peer state at end-of-epoch
type ValidatorInfoData struct {
	PublicKey  []byte
	ShardId    uint32
	List       string
	Index      uint32
	TempRating uint32
	Rating     uint32

	RewardAddress              []byte
	LeaderSuccess              uint32
	LeaderFailure              uint32
	ValidatorSuccess           uint32
	ValidatorFailure           uint32
	NumSelectedInSuccessBlocks uint32
	AccumulatedFees            *big.Int
}

// GetPublicKey returns the public key
func (vid *ValidatorInfoData) GetPublicKey() []byte {
	return vid.PublicKey
}

// GetShardId returns shard id
func (vid *ValidatorInfoData) GetShardId() uint32 {
	return vid.ShardId
}

// GetList return list the validator is in
func (vid *ValidatorInfoData) GetList() string {
	return vid.List
}

// GetIndex returns the list index
func (vid *ValidatorInfoData) GetIndex() uint32 {
	return vid.Index
}

// GetTempRating returns the temp rating
func (vid *ValidatorInfoData) GetTempRating() uint32 {
	return vid.TempRating
}

// GetRating return the current rating
func (vid *ValidatorInfoData) GetRating() uint32 {
	return vid.Rating
}

// String returns the encoded string
func (vid *ValidatorInfoData) String() string {
	return fmt.Sprintf("PK:%v, ShardId: %v, List: %v, Index:%v, TempRating:%v, Rating:%v",
		display.DisplayByteSlice(vid.PublicKey), vid.ShardId, vid.List, vid.Index, vid.TempRating, vid.Rating)
}

// IsInterfaceNil returns true if underlying object is nil
func (vid *ValidatorInfoData) IsInterfaceNil() bool {
	return vid == nil
}
