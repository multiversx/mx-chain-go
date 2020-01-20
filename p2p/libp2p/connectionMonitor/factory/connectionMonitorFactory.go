package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor"
)

type connectionMonitorFactory struct {
	reconnecter                p2p.Reconnecter
	thresholdMinConnectedPeers int
	targetCount                int
}

// NewConnectionMonitorFactory creates a new instance of connectionMonitorFactory able to create ConnectionMonitor instances
//TODO(iulian) refactor this: create a factory function, remove the struct, use a specialized argument and refactor the switch
func NewConnectionMonitorFactory(
	reconnecter p2p.Reconnecter,
	thresholdMinConnectedPeers int,
	targetCount int,
) *connectionMonitorFactory {

	return &connectionMonitorFactory{
		reconnecter:                reconnecter,
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
		targetCount:                targetCount,
	}
}

// Create creates new ConnectionMonitor instances
func (cmf *connectionMonitorFactory) Create() (ConnectionMonitor, error) {
	if check.IfNil(cmf.reconnecter) {
		return &connectionMonitor.NilConnectionMonitor{}, nil
	}

	switch recon := cmf.reconnecter.(type) {
	case p2p.ReconnecterWithPauseResumeAndWatchdog:
		return connectionMonitor.NewLibp2pConnectionMonitor(recon, cmf.thresholdMinConnectedPeers, cmf.targetCount)
	default:
		return connectionMonitor.NewLibp2pConnectionMonitorSimple(recon, cmf.thresholdMinConnectedPeers)
	}
}
