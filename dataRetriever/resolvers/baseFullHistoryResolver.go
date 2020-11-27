package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type baseFullHistoryResolver struct {
	storer storage.Storer
}

func (bfhr *baseFullHistoryResolver) getFromStorage(key []byte, epoch uint32) ([]byte, error) {
	//we just call the storer to search in the provided epoch. (it will search automatically also in the next epoch)
	return bfhr.storer.GetFromEpoch(key, epoch)
}

func (bfhr *baseFullHistoryResolver) searchFirst(key []byte) ([]byte, error) {
	return bfhr.storer.SearchFirst(key)
}
