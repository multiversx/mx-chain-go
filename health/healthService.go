package health

import (
	"context"
	"os"
	"runtime"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
)

var log = logger.GetOrCreate("health")

type healthService struct {
	config     config.HealthServiceConfig
	cancelFunc func()
	records    *records
}

func NewHealthService(config config.HealthServiceConfig) *healthService {
	log.Info("NewHealthService", "config", config)

	records := newRecords(config.NumMemoryRecordsToKeep, config.FolderPath)
	return &healthService{
		config:     config,
		cancelFunc: func() {},
		records:    records,
	}
}

func (h *healthService) Start() {
	log.Info("healthService.Start()")

	h.prepareFolder()

	ctx, cancelFunc := context.WithCancel(context.Background())
	h.cancelFunc = cancelFunc

	go h.monitorMemoryContinuously(ctx)
}

func (h *healthService) prepareFolder() {
	err := os.MkdirAll(h.config.FolderPath, os.ModePerm)
	if err != nil {
		log.Error("healthService.prepareFolder", "err", err)
	}
}

func (h *healthService) monitorMemoryContinuously(ctx context.Context) {
	for {
		if h.shouldContinueMonitoringMemory(ctx) {
			h.monitorMemory()
		} else {
			break
		}
	}

	log.Info("end of healthService.monitorMemoryContinuously()")
}

func (h *healthService) shouldContinueMonitoringMemory(ctx context.Context) bool {
	interval := time.Duration(h.config.IntervalVerifyMemoryInSeconds) * time.Second

	select {
	case <-time.After(interval):
		return true
	case <-ctx.Done():
		return false
	}
}

func (h *healthService) monitorMemory() {
	log.Trace("healthService.monitorMemory()")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	if int(stats.HeapInuse) > h.config.MemoryHighThreshold {
		record := newMemoryRecord(stats)
		h.records.addMemoryRecord(record)
	}
}

// Close stops the service
func (h *healthService) Close() error {
	h.cancelFunc()
	return nil
}
