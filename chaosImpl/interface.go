package chaosImpl

import mainFactory "github.com/multiversx/mx-chain-go/factory"

// NodeHandler -
type NodeHandler interface {
	GetCoreComponents() mainFactory.CoreComponentsHolder
	GetStatusCoreComponents() mainFactory.StatusCoreComponentsHolder
	GetCryptoComponents() mainFactory.CryptoComponentsHolder
	GetConsensusComponents() mainFactory.ConsensusComponentsHolder
	GetBootstrapComponents() mainFactory.BootstrapComponentsHolder
	GetDataComponents() mainFactory.DataComponentsHolder
	GetNetworkComponents() mainFactory.NetworkComponentsHolder
	GetProcessComponents() mainFactory.ProcessComponentsHolder
	GetStateComponents() mainFactory.StateComponentsHolder
	GetStatusComponents() mainFactory.StatusComponentsHolder
}
