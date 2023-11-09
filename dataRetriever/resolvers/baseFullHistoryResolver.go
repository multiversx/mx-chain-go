package resolvers

import (
	"github.com/multiversx/mx-chain-go/storage"
)

type baseFullHistoryResolver struct {
	storer storage.Storer
}

func (bfhr *baseFullHistoryResolver) getFromStorage(key []byte, epoch uint32) ([]byte, error) {
	//we just call the storer to search in the provided epoch. (it will search automatically also in the next epoch)
	buff, err := bfhr.storer.GetFromEpoch(key, epoch)
	if err != nil {
		// default to a search first, maximize the chance of getting recent data
		return bfhr.storer.SearchFirst(key)
	}

	return buff, err
}

func (bfhr *baseFullHistoryResolver) searchFirst(key []byte) ([]byte, error) {
	return bfhr.storer.SearchFirst(key)
}
