package storageUnit

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
)

var cummulatedSizeInBytes atomic.Counter

func monitorNewCache(sizeInBytes uint64) {
	cummulatedSizeInBytes.Add(int64(sizeInBytes))
	log.Info("NewCache", "sizeInBytes", sizeInBytes, "cummulated", core.ConvertBytes(cummulatedSizeInBytes.GetUint64()))
}
