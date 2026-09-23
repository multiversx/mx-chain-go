package sync

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

var log = logger.GetOrCreate("epochStart/bootstrap/sync")

// GetDataFromStorage searches for data from storage
func GetDataFromStorage(hash []byte, storer storage.Storer) ([]byte, error) {
	if check.IfNil(storer) {
		return nil, epochStart.ErrNilStorage
	}

	currData, err := storer.Get(hash)

	return currData, err
}

// WaitFor waits for the channel to be set or for timeout
func WaitFor(channel chan bool, waitTime time.Duration) error {
	select {
	case <-channel:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}
