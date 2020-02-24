package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor"
)

// ArgsConnectionMonitorFactory represents the argument for the connection monitor factory
type ArgsConnectionMonitorFactory struct {
	Reconnecter                p2p.Reconnecter
	ThresholdMinConnectedPeers int
	TargetCount                int
}

// NewConnectionMonitor creates a new ConnectionMonitor instance
func NewConnectionMonitor(arg ArgsConnectionMonitorFactory) (ConnectionMonitor, error) {
	if check.IfNil(arg.Reconnecter) {
		return &connectionMonitor.NilConnectionMonitor{}, nil
	}

	switch recon := arg.Reconnecter.(type) {
	case p2p.ReconnecterWithPauseResumeAndWatchdog:
		return connectionMonitor.NewLibp2pConnectionMonitor(recon, arg.ThresholdMinConnectedPeers, arg.TargetCount)
	default:
		return connectionMonitor.NewLibp2pConnectionMonitorSimple(recon, arg.ThresholdMinConnectedPeers)
	}
}
