package process

import (
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"

	chainData "github.com/multiversx/mx-chain-core-go/data"
)

// NodeHandler defines what a node handler should be able to do
type NodeHandler interface {
	GetProcessComponents() factory.ProcessComponentsHolder
	GetChainHandler() chainData.ChainHandler
	GetBroadcastMessenger() consensus.BroadcastMessenger
	GetShardCoordinator() sharding.Coordinator
	GetCryptoComponents() factory.CryptoComponentsHolder
	GetCoreComponents() factory.CoreComponentsHolder
	GetStateComponents() factory.StateComponentsHolder
	GetFacadeHandler() shared.FacadeHandler
	GetStatusCoreComponents() factory.StatusCoreComponentsHolder
	SetKeyValueForAddress(addressBytes []byte, state map[string]string) error
	SetStateForAddress(address []byte, state *dtos.AddressState) error
	GetIncomingHeaderHandler() process.IncomingHeaderSubscriber
	Close() error
	IsInterfaceNil() bool
}
