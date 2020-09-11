package workItems

import (
	"time"
)

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
}

type itemRounds struct {
	indexer    saveRounds
	roundsInfo []RoundInfo
}

// NewItemRounds will create a new instance of itemRounds
func NewItemRounds(indexer saveRounds, roundsInfo []RoundInfo) WorkItemHandler {
	return &itemRounds{
		indexer:    indexer,
		roundsInfo: roundsInfo,
	}
}

// Save will save in elasticsearch database information about rounds
func (wir *itemRounds) Save() error {
	err := wir.indexer.SaveRoundsInfo(wir.roundsInfo)
	if err != nil {
		log.Warn("itemRounds.Save", "could not index rounds info", err.Error())
		return err
	}

	return nil
}
