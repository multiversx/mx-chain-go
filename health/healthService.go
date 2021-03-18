package health

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
)

var log = logger.GetOrCreate("health")

type healthService struct {
	config                              config.HealthServiceConfig
	folder                              string
	cancelFunction                      func()
	records                             *records
	diagnosableComponents               []diagnosable
	diagnosableComponentsMutex          sync.RWMutex
	clock                               clock
	memory                              memory
	onMonitorContinuouslyBeginIteration func()
	onMonitorContinuouslyEndIteration   func()
}

// NewHealthService creates a new HealthService object
func NewHealthService(config config.HealthServiceConfig, workingDir string) *healthService {
	log.Info("NewHealthService", "config", config)

	folder := path.Join(workingDir, config.FolderPath)
	recordsObj := newRecords(config.NumMemoryUsageRecordsToKeep)

	return &healthService{
		config:                              config,
		folder:                              folder,
		cancelFunction:                      func() {},
		records:                             recordsObj,
		diagnosableComponents:               make([]diagnosable, 0),
		clock:                               &realClock{},
		memory:                              &realMemory{},
		onMonitorContinuouslyBeginIteration: func() {},
		onMonitorContinuouslyEndIteration:   func() {},
	}
}

// RegisterComponent registers a diagnosable component
func (h *healthService) RegisterComponent(component interface{}) {
	err := h.doRegisterComponent(component)
	if err != nil {
		log.Error("healthService.RegisterComponent()", "err", err, "component", fmt.Sprintf("%T", component))
	}
}

func (h *healthService) doRegisterComponent(component interface{}) error {
	asDiagnosable, ok := component.(diagnosable)
	if !ok {
		return errNotDiagnosableComponent
	}
	if check.IfNil(asDiagnosable) {
		return errNilComponent
	}

	h.diagnosableComponentsMutex.Lock()
	h.diagnosableComponents = append(h.diagnosableComponents, asDiagnosable)
	h.diagnosableComponentsMutex.Unlock()
	return nil
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
	intervalVerifyMemoryInSeconds := time.Duration(h.config.IntervalVerifyMemoryInSeconds) * time.Second
	intervalDiagnoseComponentsInSeconds := time.Duration(h.config.IntervalDiagnoseComponentsInSeconds) * time.Second
	intervalDiagnoseComponentsDeeplyInSeconds := time.Duration(h.config.IntervalDiagnoseComponentsDeeplyInSeconds) * time.Second

	chanMonitorMemory := h.clock.after(intervalVerifyMemoryInSeconds)
	chanDiagnoseComponents := h.clock.after(intervalDiagnoseComponentsInSeconds)
	chanDiagnoseComponentsDeeply := h.clock.after(intervalDiagnoseComponentsDeeplyInSeconds)

	for {
		h.onMonitorContinuouslyBeginIteration()

		select {
		case <-chanMonitorMemory:
			h.monitorMemory()
			chanMonitorMemory = h.clock.after(intervalVerifyMemoryInSeconds)
		case <-chanDiagnoseComponents:
			h.diagnoseComponents(false)
			chanDiagnoseComponents = h.clock.after(intervalDiagnoseComponentsInSeconds)
		case <-chanDiagnoseComponentsDeeply:
			h.diagnoseComponents(true)
			chanDiagnoseComponentsDeeply = h.clock.after(intervalDiagnoseComponentsDeeplyInSeconds)
		case <-ctx.Done():
			log.Info("healthService.monitorContinuously() ended")
			return
		}

		h.onMonitorContinuouslyEndIteration()
	}
}

func (h *healthService) monitorMemory() {
	stats := h.memory.getStats()

	log.Trace("healthService.monitorMemory()", "heapInUse", core.ConvertBytes(stats.HeapInuse))

	if int(stats.HeapInuse) > h.config.MemoryUsageToCreateProfiles {
		recordObj := newMemoryUsageRecord(stats, h.clock.now(), h.folder)
		h.records.addRecord(recordObj)
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
