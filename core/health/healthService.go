package health

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
)

var log = logger.GetOrCreate("health")

type healthService struct {
}

func NewHealthService(config config.HealthServiceConfig) *healthService {
	log.Info("NewHealthService", "config", config)
	return &healthService{}
}

func (h *healthService) Start() {
}
