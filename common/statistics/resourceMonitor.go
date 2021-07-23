package statistics

import (
	"context"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common/statistics/machine"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

var log = logger.GetOrCreate("common/statistics")

// ResourceMonitor outputs statistics about resources used by the binary
type ResourceMonitor struct {
	startTime     time.Time
	cancelFunc    context.CancelFunc
	generalConfig *config.Config
	pathManager   storage.PathManagerHandler
	shardId       string
}

// NewResourceMonitor creates a new ResourceMonitor instance
func NewResourceMonitor(config *config.Config, pathManager storage.PathManagerHandler, shardId string) (*ResourceMonitor, error) {
	if config == nil {
		return nil, ErrNilConfig
	}
	if check.IfNil(pathManager) {
		return nil, ErrNilPathHandler
	}

	return &ResourceMonitor{
		generalConfig: config,
		pathManager:   pathManager,
		shardId:       shardId,
		startTime:     time.Now(),
	}, nil
}

// GenerateStatistics creates a new statistic string
func (rm *ResourceMonitor) GenerateStatistics() []interface{} {
	fileDescriptors := int32(0)
	numOpenFiles := 0
	numConns := 0
	proc, err := machine.GetCurrentProcess()
	if err == nil {
		fileDescriptors, _ = proc.NumFDs()
		var openFiles []process.OpenFilesStat
		openFiles, err = proc.OpenFiles()
		if err == nil {
			numOpenFiles = len(openFiles)
		}

		var conns []net.ConnectionStat
		conns, err = proc.Connections()
		if err == nil {
			numConns = len(conns)
		}
	}

	pathManager := rm.pathManager
	generalConfig := rm.generalConfig
	shardId := rm.shardId

	trieStoragePath, mainDb := path.Split(pathManager.PathForStatic(shardId, generalConfig.AccountsTrieStorage.DB.FilePath))

	trieDbFilePath := filepath.Join(trieStoragePath, mainDb)
	evictionWaitingListDbFilePath := filepath.Join(trieStoragePath, generalConfig.EvictionWaitingList.DB.FilePath)
	snapshotsDbFilePath := filepath.Join(trieStoragePath, generalConfig.TrieSnapshotDB.FilePath)

	peerTrieStoragePath, mainDb := path.Split(pathManager.PathForStatic(shardId, generalConfig.PeerAccountsTrieStorage.DB.FilePath))

	peerTrieDbFilePath := filepath.Join(peerTrieStoragePath, mainDb)
	peerTrieEvictionWaitingListDbFilePath := filepath.Join(peerTrieStoragePath, generalConfig.EvictionWaitingList.DB.FilePath)
	peerTrieSnapshotsDbFilePath := filepath.Join(peerTrieStoragePath, generalConfig.TrieSnapshotDB.FilePath)

	stats := []interface{}{
		"uptime", time.Duration(time.Now().UnixNano() - rm.startTime.UnixNano()).Round(time.Second),
	}
	stats = append(stats, GetRuntimeStatistics()...)
	stats = append(stats, []interface{}{
		"FDs", fileDescriptors,
		"num opened files", numOpenFiles,
		"num conns", numConns,
		"accountsTrieDbMem", getDirMemSize(trieDbFilePath),
		"evictionDbMem", getDirMemSize(evictionWaitingListDbFilePath),
		"snapshotsDbMem", getDirMemSize(snapshotsDbFilePath),
		"peerTrieDbMem", getDirMemSize(peerTrieDbFilePath),
		"peerTrieEvictionDbMem", getDirMemSize(peerTrieEvictionWaitingListDbFilePath),
		"peerTrieSnapshotsDbMem", getDirMemSize(peerTrieSnapshotsDbFilePath),
	}...,
	)

	return stats
}

// GetRuntimeStatistics will return the statistics regarding the current time, memory consumption and the number of running go routines
// These return statistics can be easily output in a log line
func GetRuntimeStatistics() []interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return []interface{}{
		"timestamp", time.Now().Unix(),
		"num go", runtime.NumGoroutine(),
		"alloc", core.ConvertBytes(memStats.Alloc),
		"heap alloc", core.ConvertBytes(memStats.HeapAlloc),
		"heap idle", core.ConvertBytes(memStats.HeapIdle),
		"heap inuse", core.ConvertBytes(memStats.HeapInuse),
		"heap sys", core.ConvertBytes(memStats.HeapSys),
		"heap num objs", memStats.HeapObjects,
		"sys mem", core.ConvertBytes(memStats.Sys),
		"num GC", memStats.NumGC,
	}
}

func getDirMemSize(dir string) string {
	files, _ := ioutil.ReadDir(dir)

	size := int64(0)
	for _, f := range files {
		size += f.Size()
	}

	return core.ConvertBytes(uint64(size))
}

// SaveStatistics generates and saves statistic data on the disk
func (rm *ResourceMonitor) SaveStatistics() {
	stats := rm.GenerateStatistics()
	log.Debug("node statistics", stats...)
}

// StartMonitoring starts the monitoring process for saving statistics
func (rm *ResourceMonitor) StartMonitoring() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	rm.cancelFunc = cancelFunc
	go func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(rm.generalConfig.ResourceStats.RefreshIntervalInSec)):
				rm.SaveStatistics()
			case <-ctx.Done():
				log.Debug("closing ResourceMonitor.StartMonitoring go routine")
				return
			}
		}
	}()
}

// IsInterfaceNil returns true if underlying object is nil
func (rm *ResourceMonitor) IsInterfaceNil() bool {
	return rm == nil
}

// Close closes all underlying components
func (rm *ResourceMonitor) Close() error {
	if rm.cancelFunc != nil {
		rm.cancelFunc()
	}

	return nil
}
