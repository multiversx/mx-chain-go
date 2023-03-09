package monitor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
)

type heartbeatMessages struct {
	firstActivePid          core.PeerID
	firstActiveHeartbeat    *data.PubKeyHeartbeat
	numActivePids           uint64
	latestInactivePid       core.PeerID
	latestInactiveHeartbeat *data.PubKeyHeartbeat
	inactivePids            []core.PeerID
}

func newHeartbeatMessages() *heartbeatMessages {
	return &heartbeatMessages{}
}

func (instance *heartbeatMessages) addMessage(pid core.PeerID, message *data.PubKeyHeartbeat) {
	if message.IsActive {
		instance.handleActiveMessage(pid, message)
		return
	}

	instance.handleInactiveMessage(pid, message)
}

func (instance *heartbeatMessages) handleActiveMessage(pid core.PeerID, message *data.PubKeyHeartbeat) {
	instance.numActivePids++

	if len(instance.firstActivePid) == 0 {
		instance.firstActivePid = pid
		instance.firstActiveHeartbeat = message
		return
	}

	if string(instance.firstActivePid) > string(pid) {
		instance.firstActivePid = pid
		instance.firstActiveHeartbeat = message
	}
}

func (instance *heartbeatMessages) handleInactiveMessage(pid core.PeerID, message *data.PubKeyHeartbeat) {
	instance.inactivePids = append(instance.inactivePids, pid)

	if instance.latestInactiveHeartbeat == nil {
		instance.latestInactivePid = pid
		instance.latestInactiveHeartbeat = message
		return
	}

	if message.TimeStamp.Unix() > instance.latestInactiveHeartbeat.TimeStamp.Unix() {
		instance.latestInactivePid = pid
		instance.latestInactiveHeartbeat = message
	}
}

func (instance *heartbeatMessages) getHeartbeat() (*data.PubKeyHeartbeat, error) {
	if instance.firstActiveHeartbeat != nil {
		return instance.firstActiveHeartbeat, nil
	}
	if instance.latestInactiveHeartbeat != nil {
		return instance.latestInactiveHeartbeat, nil
	}
	return nil, errEmptyHeartbeatMessagesInstance
}

func (instance *heartbeatMessages) getInactivePids() []core.PeerID {
	if instance.firstActiveHeartbeat != nil {
		// at least one active heartbeat, return all pids containing inactive messages
		return instance.inactivePids
	}

	result := make([]core.PeerID, 0, len(instance.inactivePids))
	for _, pid := range instance.inactivePids {
		if pid == instance.latestInactivePid {
			continue
		}

		result = append(result, pid)
	}

	return result
}
