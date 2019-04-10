package statistics

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
)

var errNilFileToWriteStats = errors.New("nil file to write statistics")

// ResourceMonitor outputs statistics about resources used by the binary
type ResourceMonitor struct {
	startTime time.Time
	file      *os.File
	mutFile   sync.RWMutex
}

// NewResourceMonitor creates a new ResourceMonitor instance
func NewResourceMonitor(file *os.File) (*ResourceMonitor, error) {
	if file == nil {
		return nil, errNilFileToWriteStats
	}

	return &ResourceMonitor{
		startTime: time.Now(),
		file:      file,
	}, nil
}

// GenerateStatistics creates a new statistic string
func (rm *ResourceMonitor) GenerateStatistics() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return fmt.Sprintf("timestamp: %d, uptime: %v, num go: %d, go mem: %s, sys mem: %s, total mem: %s, num GC: %d\n",
		time.Now().Unix(),
		time.Duration(time.Now().UnixNano()-rm.startTime.UnixNano()).Round(time.Second),
		runtime.NumGoroutine(),
		core.ConvertBytes(memStats.Alloc),
		core.ConvertBytes(memStats.Sys),
		core.ConvertBytes(memStats.TotalAlloc),
		memStats.NumGC,
	)
}

// SaveStatistics generates and saves statistic data on the disk
func (rm *ResourceMonitor) SaveStatistics() error {
	rm.mutFile.RLock()
	defer rm.mutFile.RUnlock()
	if rm.file == nil {
		return errNilFileToWriteStats
	}

	stats := rm.GenerateStatistics()
	_, err := rm.file.WriteString(stats)
	if err != nil {
		return err
	}

	err = rm.file.Sync()
	if err != nil {
		return err
	}

	return nil
}

// Close closes the file used for statistics
func (rm *ResourceMonitor) Close() error {
	rm.mutFile.Lock()
	defer rm.mutFile.Unlock()

	err := rm.file.Close()
	rm.file = nil
	return err
}
