package heartbeat

import (
	"sync"

	"github.com/multiversx/mx-chain-go/heartbeat/data"
)

type heartbeatMonitor struct {
	heartbeats []data.PubKeyHeartbeat
	mutex      sync.RWMutex
}

// NewHeartbeatMonitor will create a new instance of heartbeat monitor
func NewHeartbeatMonitor() *heartbeatMonitor {
	return &heartbeatMonitor{
		heartbeats: make([]data.PubKeyHeartbeat, 0),
	}
}

// GetHeartbeats will return the heartbeats
func (hm *heartbeatMonitor) GetHeartbeats() []data.PubKeyHeartbeat {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	return hm.heartbeats
}

// SetHeartbeats will set the provided heartbeats
func (hm *heartbeatMonitor) SetHeartbeats(heartbeats []data.PubKeyHeartbeat) {
	hm.mutex.Lock()
	hm.heartbeats = heartbeats
	hm.mutex.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (hm *heartbeatMonitor) IsInterfaceNil() bool {
	return nil == hm
}
