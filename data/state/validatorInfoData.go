package state

import "fmt"

type ValidatorInfoData struct {
	PublicKey  []byte
	ShardId    uint32
	List       string
	Index      uint32
	TempRating uint32
	Rating     uint32
}

func NewValidatorInfoData(publicKey []byte, shardId uint32, list string, index uint32, tempRating uint32, rating uint32) (*ValidatorInfoData, error) {
	return &ValidatorInfoData{
		PublicKey:  publicKey,
		ShardId:    shardId,
		List:       list,
		Index:      index,
		TempRating: tempRating,
		Rating:     rating,
	}, nil
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
