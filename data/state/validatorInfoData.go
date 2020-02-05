package state

import "fmt"

type ValidatorInfoData struct {
	publicKey  []byte
	shardId    uint32
	list       string
	index      uint32
	tempRating uint32
	rating     uint32
}

func NewValidatorInfoData(publicKey []byte, shardId uint32, list string, index uint32, tempRating uint32, rating uint32) *ValidatorInfoData {
	return &ValidatorInfoData{
		publicKey:  publicKey,
		shardId:    shardId,
		list:       list,
		index:      index,
		tempRating: tempRating,
		rating:     rating,
	}
}

func (vid *ValidatorInfoData) PublicKey() []byte {
	return vid.publicKey
}

func (vid *ValidatorInfoData) ShardId() uint32 {
	return vid.shardId
}

func (vid *ValidatorInfoData) List() string {
	return vid.list
}

func (vid *ValidatorInfoData) Index() uint32 {
	return vid.index
}

func (vid *ValidatorInfoData) TempRating() uint32 {
	return vid.tempRating
}

func (vid *ValidatorInfoData) Rating() uint32 {
	return vid.rating
}

func (vid *ValidatorInfoData) String() string {
	return fmt.Sprintf("PK:%v, ShardId: %v, List: %v, Index:%v, TempRating:%v, Rating:%v",
		vid.publicKey, vid.shardId, vid.list, vid.index, vid.tempRating, vid.rating)
}
