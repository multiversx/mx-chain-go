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
	"github.com/ElrondNetwork/elrond-go/core"
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
	go h.monitorContinuously(ctx)
}

func (h *healthService) prepareFolder() {
	err := os.MkdirAll(h.config.FolderPath, os.ModePerm)
	if err != nil {
		log.Error("healthService.prepareFolder", "err", err)
	}
}

func (h *healthService) monitorContinuously(ctx context.Context) {
	for i := 0; h.shouldContinueInfiniteLoop(ctx); i++ {
		shouldMonitorMemory := i%h.config.IntervalVerifyMemoryInSeconds == 0
		shouldDiagnoseComponents := i%h.config.IntervalDiagnoseComponentsInSeconds == 0
		shouldDiagnoseComponentsDeeply := i%h.config.IntervalDiagnoseComponentsDeeplyInSeconds == 0

		if shouldMonitorMemory {
			h.monitorMemory()
		}
		if shouldDiagnoseComponents {
			h.diagnoseComponents(false)
		}
		if shouldDiagnoseComponentsDeeply {
			h.diagnoseComponents(true)
		}
	}

	log.Info("healthService.monitorContinuously() ended")
}

func (h *healthService) shouldContinueInfiniteLoop(ctx context.Context) bool {
	select {
	case <-time.After(time.Second):
		return true
	case <-ctx.Done():
		return false
	}
}

func (h *healthService) monitorMemory() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	log.Trace("healthService.monitorMemory()", "heapInUse", core.ConvertBytes(stats.HeapInuse))

	if int(stats.HeapInuse) > h.config.MemoryToCreateProfiles {
		record := newMemoryRecord(stats)
		h.records.addMemoryRecord(record)
	}
}

func (h *healthService) diagnoseComponents(deep bool) {
	log.Trace("healthService.diagnoseComponents()", "deep", deep)

	h.diagnosableComponentsMutex.RLock()
	defer h.diagnosableComponentsMutex.RUnlock()

	for _, component := range h.diagnosableComponents {
		component.Diagnose(deep)
	}
}

func (h *healthService) RegisterComponent(component interface{}) {
	log.Debug("healthService.RegisterComponent()", "component", fmt.Sprintf("%T", component))

	h.diagnosableComponentsMutex.Lock()
	defer h.diagnosableComponentsMutex.Unlock()

	asDiagnosable, ok := component.(diagnosable)
	if !ok {
		log.Error("healthService.RegisterComponent(): not diagnosable", "component", fmt.Sprintf("%T", component))
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
