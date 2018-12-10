package facade

import (
	"sync"
	"time"
)

type Facade interface {
	CreateNode(maxAllowedPeers, port int)
	StartNode(initialAddresses []string) error
	StartNTP(clockSyncPeriod int)
	WaitForStartTime(t time.Time)
	StartBackgroundServices(wg *sync.WaitGroup)
}
