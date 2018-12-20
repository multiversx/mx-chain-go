package facade

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

type Facade interface {
	CreateNode(maxAllowedPeers, port int, initialNodeAddresses []string)
	StartNode(initialAddresses []string) error
	StopNode() error
	StartNTP(clockSyncPeriod int)
	WaitForStartTime(t time.Time)
	StartBackgroundServices(wg *sync.WaitGroup)
	SetLogger(logger *logger.Logger)
	IsNodeRunning() bool
}
