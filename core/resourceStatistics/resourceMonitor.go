package resourceStatistics

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const defaultStatPath = "stats"

var errNilFileToWriteStats = errors.New("nil file to write statistics")

// ResourceMonitor outputs statistics about resources used by the binary
type ResourceMonitor struct {
	startTime time.Time
	file      *os.File
	mutFile   sync.RWMutex
}

// NewResourceMonitor creates a new ResourceMonitor instance
func NewResourceMonitor(fileId string) (*ResourceMonitor, error) {
	rm := &ResourceMonitor{
		startTime: time.Now(),
	}

	var err error
	rm.file, err = newStatsFile(fileId)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

// GenerateStatistics creates a new statistic string
func (rm *ResourceMonitor) GenerateStatistics() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return fmt.Sprintf("timestamp: %d, uptime: %v, num go: %d, go mem: %s, sys mem: %s, total mem: %s, num GC: %d\n",
		time.Now().Unix(),
		time.Duration(time.Now().UnixNano()-rm.startTime.UnixNano()).Round(time.Second),
		runtime.NumGoroutine(),
		convertBytes(memStats.Alloc),
		convertBytes(memStats.Sys),
		convertBytes(memStats.TotalAlloc),
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
	_, err = rm.file.WriteString("")
	if err != nil {
		return err
	}

	err = rm.file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func convertBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f kiB", float64(bytes)/1024.0)
	}
	if bytes < 1024*1024*1025 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024.0/1024.0)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/1024.0/1024.0/1024.0)
}

// newStatsFile returns new output for the application logger.
func newStatsFile(fileId string) (*os.File, error) {
	absPath, err := filepath.Abs(defaultStatPath)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("stat_%s_%s.txt", fileId, time.Now().Format("2006-02-01"))
	fileWithPath := filepath.Join(absPath, filename)

	_, err = os.Stat(fileWithPath)
	if !os.IsNotExist(err) {
		//save the existing file with another name
		filenameOld := fmt.Sprintf("stat_%s_%s.old", fileId, time.Now().Format("2006-02-01-15-04-05"))
		oldFileWithPath := filepath.Join(absPath, filenameOld)

		err = os.Rename(fileWithPath, oldFileWithPath)
		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(fileWithPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0666)
}

// Close closes the file used for statistics
func (rm *ResourceMonitor) Close() error {
	rm.mutFile.Lock()
	defer rm.mutFile.Unlock()

	err := rm.file.Close()
	rm.file = nil
	return err
}
