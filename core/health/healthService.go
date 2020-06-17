package health

import (
	"context"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
)

var log = logger.GetOrCreate("health")

type healthService struct {
	config     config.HealthServiceConfig
	cancelFunc func()
}

func NewHealthService(config config.HealthServiceConfig) *healthService {
	log.Info("NewHealthService", "config", config)

	return &healthService{
		config:     config,
		cancelFunc: func() {},
	}
}

func (h *healthService) Start() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	h.cancelFunc = cancelFunc

	go h.monitorMemoryContinuously(ctx)
}

func (h *healthService) monitorMemoryContinuously(ctx context.Context) {
	for {
		if h.shouldContinueMonitoringMemory(ctx) {
			h.monitorMemory()
		} else {
			break
		}
	}

	log.Info("ending monitorMemoryContinuously")
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
}

// Close stops the service
func (h *healthService) Close() error {
	h.cancelFunc()
	return nil
}
