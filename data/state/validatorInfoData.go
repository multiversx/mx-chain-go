package state

import (
	"fmt"
	"math/big"
)

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

func (vid *ValidatorInfoData) GetPublicKey() []byte {
	return vid.PublicKey
}

func (vid *ValidatorInfoData) GetShardId() uint32 {
	return vid.ShardId
}

func (vid *ValidatorInfoData) GetList() string {
	return vid.List
}

func (vid *ValidatorInfoData) GetIndex() uint32 {
	return vid.TempRating
}

func (vid *ValidatorInfoData) GetTempRating() uint32 {
	return vid.TempRating
}

func (vid *ValidatorInfoData) GetRating() uint32 {
	return vid.Rating
}

func (vid *ValidatorInfoData) String() string {
	return fmt.Sprintf("PK:%v, ShardId: %v, List: %v, Index:%v, TempRating:%v, Rating:%v",
		vid.PublicKey, vid.ShardId, vid.List, vid.Index, vid.TempRating, vid.Rating)
}

// IsInterfaceNil returns true if underlying object is nil
func (vid *ValidatorInfoData) IsInterfaceNil() bool {
	return vid == nil
}
