package workItems

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"publicKey"`
	Rating    float32 `json:"rating"`
}

type itemRating struct {
	indexer    saveRatingIndexer
	indexID    string
	infoRating []ValidatorRatingInfo
}

// NewItemRating will create a new instance of itemRating
func NewItemRating(indexer saveRatingIndexer, indexID string, infoRating []ValidatorRatingInfo) WorkItemHandler {
	return &itemRating{
		indexer:    indexer,
		indexID:    indexID,
		infoRating: infoRating,
	}
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
