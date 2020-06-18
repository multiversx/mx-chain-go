package health

import (
	"context"
	"fmt"
	"os"
	"path"
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
	folder                     string
	cancelFunction             func()
	records                    *records
	diagnosableComponents      []diagnosable
	diagnosableComponentsMutex sync.RWMutex
}

// NewHealthService creates a new HealthService object
func NewHealthService(config config.HealthServiceConfig, workingDir string) *healthService {
	log.Info("NewHealthService", "config", config)

	folder := path.Join(workingDir, config.FolderPath)
	records := newRecords(config.NumMemoryRecordsToKeep)

	return &healthService{
		config:                config,
		folder:                folder,
		cancelFunction:        func() {},
		records:               records,
		diagnosableComponents: make([]diagnosable, 0),
	}
}

// RegisterComponent registers a diagnosable component
func (h *healthService) RegisterComponent(component interface{}) {
	h.diagnosableComponentsMutex.Lock()
	defer h.diagnosableComponentsMutex.Unlock()

	asDiagnosable, ok := component.(diagnosable)
	if !ok {
		log.Error("healthService.RegisterComponent(): not diagnosable", "component", fmt.Sprintf("%T", component))
		return
	}

	h.diagnosableComponents = append(h.diagnosableComponents, asDiagnosable)
}

// Start starts the health service
func (h *healthService) Start() {
	log.Info("healthService.Start()")

	h.prepareFolder()
	ctx := h.setupCancellation()
	go h.monitorContinuously(ctx)
}

func (h *healthService) prepareFolder() {
	err := os.MkdirAll(h.folder, os.ModePerm)
	if err != nil {
		log.Error("healthService.prepareFolder", "err", err)
	}
}

func (h *healthService) setupCancellation() context.Context {
	ctx, cancelFunc := context.WithCancel(context.Background())
	h.cancelFunction = cancelFunc
	return ctx
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
		record := newMemoryRecord(stats, h.folder)
		h.records.addRecord(record)
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

// Close stops the service
func (h *healthService) Close() error {
	h.cancelFunction()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *healthService) IsInterfaceNil() bool {
	return h == nil
}
