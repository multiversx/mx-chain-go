package sync

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

func GetDataFromStorage(hash []byte, storer update.HistoryStorer, syncingEpoch uint32) ([]byte, error) {
	currData, err := storer.Get(hash)
	if err != nil {
		currData, err = storer.GetFromEpoch(hash, syncingEpoch)
		if err != nil {
			currData, err = storer.GetFromEpoch(hash, syncingEpoch-1)
		}
	}

	return currData, err
}

func WaitFor(channel chan bool, waitTime time.Duration) error {
	select {
	case <-channel:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}
