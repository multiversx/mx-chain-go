package status

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/epochStart"
	outportDriverFactory "github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/p2p"
)

// EpochStartEventHandler -
func (pc *statusComponents) EpochStartEventHandler() epochStart.ActionHandler {
	return pc.epochStartEventHandler()
}

// ComputeNumConnectedPeers -
func ComputeNumConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	netMessenger p2p.Messenger,
) {
	computeNumConnectedPeers(appStatusHandler, netMessenger)
}

// ComputeConnectedPeers -
func ComputeConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	netMessenger p2p.Messenger,
) {
	computeConnectedPeers(appStatusHandler, netMessenger)
}

// MakeHostDriversArgs -
func (scf *statusComponentsFactory) MakeHostDriversArgs() ([]outportDriverFactory.ArgsHostDriverFactory, error) {
	return scf.makeHostDriversArgs()
}
