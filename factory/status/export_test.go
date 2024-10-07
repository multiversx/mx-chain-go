package status

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/outport"
	outportDriverFactory "github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/p2p"
)

// ComputeNumConnectedPeers -
func ComputeNumConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	netMessenger p2p.Messenger,
	suffix string,
) {
	computeNumConnectedPeers(appStatusHandler, netMessenger, suffix)
}

// ComputeConnectedPeers -
func ComputeConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	netMessenger p2p.Messenger,
	suffix string,
) {
	computeConnectedPeers(appStatusHandler, netMessenger, suffix)
}

// MakeHostDriversArgs -
func (scf *statusComponentsFactory) MakeHostDriversArgs() ([]outportDriverFactory.ArgsHostDriverFactory, error) {
	return scf.makeHostDriversArgs()
}

// EpochStartEventHandler -
func (pc *statusComponents) OutportHandler() outport.OutportHandler {
	return pc.outportHandler
}
