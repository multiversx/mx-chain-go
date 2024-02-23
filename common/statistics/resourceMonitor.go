package statistics

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/statistics/machine"
	"github.com/multiversx/mx-chain-go/common/statistics/osLevel"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

const minRefreshTimeInSec = 1

var log = logger.GetOrCreate("common/statistics")

// resourceMonitor outputs statistics about resources used by the binary
type resourceMonitor struct {
	startTime     time.Time
	cancelFunc    context.CancelFunc
	netStats      NetworkStatisticsProvider
	generalConfig config.Config
}

// NewResourceMonitor creates a new ResourceMonitor instance
func NewResourceMonitor(config config.Config, netStats NetworkStatisticsProvider) (*resourceMonitor, error) {
	if check.IfNil(netStats) {
		return nil, ErrNilNetworkStatisticsProvider
	}
	if config.ResourceStats.RefreshIntervalInSec < minRefreshTimeInSec {
		return nil, fmt.Errorf("%w, minimum: %d, provided: %d", ErrInvalidRefreshIntervalValue, minRefreshTimeInSec, config.ResourceStats.RefreshIntervalInSec)
	}

	memoryString := "no memory information"
	vms, err := mem.VirtualMemory()
	if err == nil {
		memoryString = core.ConvertBytes(vms.Total)
	}

	numCores := -1
	rawCpuInfo, err := cpu.Info()
	if err == nil {
		numCores = len(rawCpuInfo)
	}

	log.Debug("newResourceMonitor", "numCores", numCores, "memory", memoryString)

	return &resourceMonitor{
		generalConfig: config,
		startTime:     time.Now(),
		netStats:      netStats,
	}, nil
}

// GenerateStatistics creates a new statistic string
func (rm *resourceMonitor) GenerateStatistics() []interface{} {
	stats := []interface{}{
		"uptime", time.Duration(time.Now().UnixNano() - rm.startTime.UnixNano()).Round(time.Second),
	}

	stats = append(stats, GetRuntimeStatistics()...)
	stats = append(stats, rm.generateProcessStatistics()...)
	stats = append(stats, rm.generateNetworkStatistics()...)

	return stats
}

func (rm *resourceMonitor) generateProcessStatistics() []interface{} {
	fileDescriptors := int32(0)
	numOpenFiles := 0
	numConns := 0
	cpuPercent := 0.0
	ioStatsString := "no io stats info"
	cpuPercentString := "no cpu percent info"
	loadAverageString := "no load average info"

	proc, err := machine.GetCurrentProcess()
	if err == nil {
		fileDescriptors, _ = proc.NumFDs()
		var openFiles []process.OpenFilesStat
		openFiles, err = proc.OpenFiles()
		if err == nil {
			numOpenFiles = len(openFiles)
		}

		loadAvg, errLoadAvg := load.Avg()
		if errLoadAvg == nil {
			loadAverageString = fmt.Sprintf("{avg1min: %.2f, avg5min: %.2f, avg15min: %.2f}",
				loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
		}

		cpuPercent, err = proc.CPUPercent()
		if err == nil {
			cpuPercentString = fmt.Sprintf("%.2f", cpuPercent)
		}

		ioStats, errIoCounters := proc.IOCounters()
		if errIoCounters == nil {
			ioStatsString = fmt.Sprintf("{readCount: %d, writeCount: %d, readBytes: %s, writeBytes: %s}",
				ioStats.ReadCount,
				ioStats.WriteCount,
				core.ConvertBytes(ioStats.ReadBytes),
				core.ConvertBytes(ioStats.WriteBytes))
		}

		var conns []net.ConnectionStat
		conns, err = proc.Connections()
		if err == nil {
			numConns = len(conns)
		}
	}

	return []interface{}{
		"FDs", fileDescriptors,
		"num opened files", numOpenFiles,
		"num conns", numConns,
		"cpuPercent", cpuPercentString,
		"cpuLoadAveragePercent", loadAverageString,
		"ioStatsString", ioStatsString,
	}
}

func (rm *resourceMonitor) generateNetworkStatistics() []interface{} {
	netStatsString := fmt.Sprintf("{total received: %s, total sent: %s}",
		rm.netStats.TotalReceivedInCurrentEpoch(),
		rm.netStats.TotalSentInCurrentEpoch(),
	)

	return []interface{}{
		"host network data size in epoch", netStatsString,
	}
}

// GetRuntimeStatistics will return the statistics regarding the current time, memory consumption and the number of running go routines
// These return statistics can be easily output in a log line
func GetRuntimeStatistics() []interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	osLevelMetrics := make([]interface{}, 0)
	osLevelMemStats, err := osLevel.ReadCurrentMemStats()
	if err == nil {
		osLevelMetrics = append(osLevelMetrics, "os level stats")
		osLevelMetrics = append(osLevelMetrics, "{"+osLevelMemStats.String()+"}")
	}

	statistics := []interface{}{
		"timestamp", time.Now().Unix(),
		"num go", runtime.NumGoroutine(),
		"heap alloc", core.ConvertBytes(memStats.HeapAlloc),
		"heap idle", core.ConvertBytes(memStats.HeapIdle),
		"heap inuse", core.ConvertBytes(memStats.HeapInuse),
		"heap sys", core.ConvertBytes(memStats.HeapSys),
		"heap num objs", memStats.HeapObjects,
		"sys mem", core.ConvertBytes(memStats.Sys),
		"num GC", memStats.NumGC,
	}
	statistics = append(statistics, osLevelMetrics...)

	return statistics
}

// LogStatistics generates and saves the statistic data in the logs
func (rm *resourceMonitor) LogStatistics() {
	stats := rm.GenerateStatistics()
	log.Debug("node statistics", stats...)
}

// StartMonitoring starts the monitoring process for saving statistics
func (rm *resourceMonitor) StartMonitoring() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	rm.cancelFunc = cancelFunc
	return
	refreshTime := time.Second * time.Duration(rm.generalConfig.ResourceStats.RefreshIntervalInSec)
	timer := time.NewTimer(refreshTime)
	defer timer.Stop()

	go func() {
		for {
			rm.LogStatistics()
			timer.Reset(refreshTime)

			select {
			case <-timer.C:
			case <-ctx.Done():
				log.Debug("closing ResourceMonitor.StartMonitoring go routine")
				return
			}
		}
	}()
}

// IsInterfaceNil returns true if underlying object is nil
func (rm *resourceMonitor) IsInterfaceNil() bool {
	return rm == nil
}

// Close closes all underlying components
func (rm *resourceMonitor) Close() error {
	if rm.cancelFunc != nil {
		rm.cancelFunc()
	}

	return nil
}
