package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
)

const (
	// BootstrapComponentsName is the bootstrap components identifier
	BootstrapComponentsName = "managedBootstrapComponents"
	// ConsensusComponentsName is the consensus components identifier
	ConsensusComponentsName = "managedConsensusComponents"
	// CoreComponentsName is the core components identifier
	CoreComponentsName = "managedCoreComponents"
	// StatusCoreComponentsName is the status core components identifier
	StatusCoreComponentsName = "managedStatusCoreComponents"
	// CryptoComponentsName is the crypto components identifier
	CryptoComponentsName = "managedCryptoComponents"
	// DataComponentsName is the data components identifier
	DataComponentsName = "managedDataComponents"
	// HeartbeatV2ComponentsName is the heartbeat V2 components identifier
	HeartbeatV2ComponentsName = "managedHeartbeatV2Components"
	// NetworkComponentsName is the network components identifier
	NetworkComponentsName = "managedNetworkComponents"
	// ProcessComponentsName is the process components identifier
	ProcessComponentsName = "managedProcessComponents"
	// StateComponentsName is the state components identifier
	StateComponentsName = "managedStateComponents"
	// StatusComponentsName is the status components identifier
	StatusComponentsName = "managedStatusComponents"
	// RunTypeComponentsName is the runType components identifier
	RunTypeComponentsName = "managedRunTypeComponents"
)

// ArgsEpochStartTrigger is a struct placeholder for arguments needed to create an epoch start trigger
type ArgsEpochStartTrigger struct {
	RequestHandler             epochStart.RequestHandler
	CoreData                   CoreComponentsHolder
	BootstrapComponents        BootstrapComponentsHolder
	DataComps                  DataComponentsHolder
	StatusCoreComponentsHolder StatusCoreComponentsHolder
	RunTypeComponentsHolder    RunTypeComponentsHolder
	Config                     config.Config
}
