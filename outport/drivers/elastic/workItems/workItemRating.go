package workItems

import (
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

type itemRating struct {
	indexer    saveRatingIndexer
	indexID    string
	infoRating []types.ValidatorRatingInfo
}

// NewItemRating will create a new instance of itemRating
func NewItemRating(indexer saveRatingIndexer, indexID string, infoRating []types.ValidatorRatingInfo) WorkItemHandler {
	return &itemRating{
		indexer:    indexer,
		indexID:    indexID,
		infoRating: infoRating,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wir *itemRating) IsInterfaceNil() bool {
	return wir == nil
}

// Save will save validators rating in elasticsearch database
func (wir *itemRating) Save() error {
	err := wir.indexer.SaveValidatorsRating(wir.indexID, wir.infoRating)
	if err != nil {
		log.Warn("itemRating.Save", "could not index validators rating", err.Error())
		return err
	}

	return nil
}
