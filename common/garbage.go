package common

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var mutRunSnapshot = sync.RWMutex{}
var canRunSnapshot = true
var log = logger.GetOrCreate("common")

func CanStartSnapshot() {
	mutRunSnapshot.Lock()
	defer mutRunSnapshot.Unlock()

	log.Debug("can start snapshot")
	canRunSnapshot = true
}

func PreventSnapshots() {
	mutRunSnapshot.Lock()
	defer mutRunSnapshot.Unlock()

	log.Debug("prevent snapshot")
	canRunSnapshot = false
}

func ShouldContinueSnapshot() bool {
	mutRunSnapshot.RLock()
	defer mutRunSnapshot.RUnlock()

	return canRunSnapshot
}
