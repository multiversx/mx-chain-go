package monitor

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

type heartbeatMessages struct {
	activePids         []core.PeerID
	activeHeartbeats   map[core.PeerID]*data.PubKeyHeartbeat
	inactiveHeartbeats map[core.PeerID]*data.PubKeyHeartbeat
}

func newHeartbeatMessages() *heartbeatMessages {
	return &heartbeatMessages{
		activePids:         make([]core.PeerID, 0),
		activeHeartbeats:   make(map[core.PeerID]*data.PubKeyHeartbeat),
		inactiveHeartbeats: make(map[core.PeerID]*data.PubKeyHeartbeat),
	}
}

func (instance *heartbeatMessages) addMessage(pid core.PeerID, message *data.PubKeyHeartbeat) {
	if message.IsActive {
		instance.activePids = append(instance.activePids, pid)
		instance.activeHeartbeats[pid] = message

		return
	}

	instance.inactiveHeartbeats[pid] = message
}

func (instance *heartbeatMessages) getHeartbeat() (*data.PubKeyHeartbeat, error) {
	if len(instance.activeHeartbeats) > 0 {
		return instance.getFirstActive()
	}
	if len(instance.inactiveHeartbeats) > 0 {
		for _, inactive := range instance.inactiveHeartbeats {
			return inactive, nil
		}
	}
	return nil, errEmptyHeartbeatMessagesInstance
}

func (instance *heartbeatMessages) getFirstActive() (*data.PubKeyHeartbeat, error) {
	sort.Slice(instance.activePids, func(i, j int) bool {
		return string(instance.activePids[i]) < string(instance.activePids[j])
	})

	if len(instance.activePids) == 0 {
		return nil, errInconsistentActivePidsList
	}

	message, found := instance.activeHeartbeats[instance.activePids[0]]
	if !found {
		return nil, errInconsistentActiveMap
	}

	return message, nil
}
