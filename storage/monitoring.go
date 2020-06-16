package storage

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
)

var log = logger.GetOrCreate("storage")

var cummulatedSizeInBytes atomic.Counter

func MonitorNewCache(tag string, sizeInBytes uint64) {
	cummulatedSizeInBytes.Add(int64(sizeInBytes))
	log.Debug("MonitorNewCache", "name", tag, "capacity", core.ConvertBytes(sizeInBytes), "cummulated", core.ConvertBytes(cummulatedSizeInBytes.GetUint64()))
}
