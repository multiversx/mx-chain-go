package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type baseFullHistoryResolver struct {
	storer storage.Storer
}

func (bfhr *baseFullHistoryResolver) getFromStorage(key []byte, epoch uint32) ([]byte, error) {
	//we first try to find it in specified epoch
	buff, err := bfhr.storer.GetFromEpoch(key, epoch)
	if err == nil {
		return buff, nil
	}

	//there might be an epoch change edge-case and worth searching in the next epoch
	return bfhr.storer.GetFromEpoch(key, epoch+1)
}

func (bfhr *baseFullHistoryResolver) searchFirst(key []byte) ([]byte, error) {
	return bfhr.storer.SearchFirst(key)
}
