package health

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
)

var log = logger.GetOrCreate("health")

type healthService struct {
	config                     config.HealthServiceConfig
	cancelFunction             func()
	records                    *records
	diagnosableComponents      []diagnosable
	diagnosableComponentsMutex sync.RWMutex
}

func NewHealthService(config config.HealthServiceConfig) *healthService {
	log.Info("NewHealthService", "config", config)

	records := newRecords(config.NumMemoryRecordsToKeep, config.FolderPath)
	return &healthService{
		config:                config,
		cancelFunction:        func() {},
		records:               records,
		diagnosableComponents: make([]diagnosable, 0),
	}
}

func (h *healthService) Start() {
	log.Info("healthService.Start()")

	h.prepareFolder()

	ctx, cancelFunc := context.WithCancel(context.Background())
	h.cancelFunction = cancelFunc

	go h.monitorMemoryContinuously(ctx)
	go h.diagnoseComponentsContinuously(ctx)
}

func (h *healthService) prepareFolder() {
	err := os.MkdirAll(h.config.FolderPath, os.ModePerm)
	if err != nil {
		log.Error("healthService.prepareFolder", "err", err)
	}
}

func (h *healthService) monitorMemoryContinuously(ctx context.Context) {
	afterSeconds := h.config.IntervalVerifyMemoryInSeconds
	for {
		if h.shouldContinueInfiniteLoop(ctx, afterSeconds) {
			h.monitorMemory()
		} else {
			break
		}
	}

	log.Info("healthService.monitorMemoryContinuously() ended")
}

func (h *healthService) shouldContinueInfiniteLoop(ctx context.Context, afterSeconds int) bool {
	interval := time.Duration(afterSeconds) * time.Second

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

	if int(stats.HeapInuse) > h.config.MemoryToCreateProfiles {
		record := newMemoryRecord(stats)
		h.records.addMemoryRecord(record)
	}
}

func (h *healthService) diagnoseComponentsContinuously(ctx context.Context) {
	afterSeconds := h.config.IntervalDiagnoseComponentsInSeconds
	for {
		if h.shouldContinueInfiniteLoop(ctx, afterSeconds) {
			h.diagnoseComponents()
		} else {
			break
		}
	}

	log.Info("healthService.RegisterComponentsContinuously() ended")
}

func (h *healthService) diagnoseComponents() {
	h.diagnosableComponentsMutex.RLock()
	defer h.diagnosableComponentsMutex.RUnlock()

	for _, component := range h.diagnosableComponents {
		log.Debug("healthService.diagnoseComponent()", "component", fmt.Sprintf("%T", component))
		component.Diagnose()
	}
}

func (h *healthService) RegisterComponent(component interface{}) {
	h.diagnosableComponentsMutex.Lock()
	defer h.diagnosableComponentsMutex.Unlock()

	asDiagnosable, ok := component.(diagnosable)
	if !ok {
		return
	}

	h.diagnosableComponents = append(h.diagnosableComponents, asDiagnosable)
}

// Close stops the service
func (h *healthService) Close() error {
	h.cancelFunction()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *healthService) IsInterfaceNil() bool {
	return h == nil
}
