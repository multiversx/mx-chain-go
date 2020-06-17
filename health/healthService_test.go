package health

import (
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logger.SetLogLevel("health:TRACE")
}

func TestNewHealthService(t *testing.T) {
	h := NewHealthService(config.HealthServiceConfig{})
	require.NotNil(t, h)
}

func TestHealthService_StartThenClose(t *testing.T) {
	h := NewHealthService(config.HealthServiceConfig{})
	require.NotNil(t, h)

	h.Start()
	_ = h.Close()

	time.Sleep(1 * time.Second)
}
