package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/sharding"
)

// NodeHandler defines what a node handler should be able to do
type NodeHandler interface {
	GetProcessComponents() factory.ProcessComponentsHolder
	GetChainHandler() chainData.ChainHandler
	GetBroadcastMessenger() consensus.BroadcastMessenger
	GetShardCoordinator() sharding.Coordinator
	GetCryptoComponents() factory.CryptoComponentsHolder
	GetCoreComponents() factory.CoreComponentsHolder
	GetDataComponents() factory.DataComponentsHolder
	GetStateComponents() factory.StateComponentsHolder
	GetFacadeHandler() shared.FacadeHandler
	GetStatusCoreComponents() factory.StatusCoreComponentsHolder
	GetNetworkComponents() factory.NetworkComponentsHolder
	SetKeyValueForAddress(addressBytes []byte, state map[string]string) error
	SetStateForAddress(address []byte, state *dtos.AddressState) error
	RemoveAccount(address []byte) error
	ForceChangeOfEpoch() error
	GetBasePeers() map[uint32]core.PeerID
	SetBasePeers(basePeers map[uint32]core.PeerID)
	Close() error
	IsInterfaceNil() bool
}
