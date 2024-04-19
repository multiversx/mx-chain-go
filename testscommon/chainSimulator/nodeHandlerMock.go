package chainSimulator

import (
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/sharding"
)

// NodeHandlerMock -
type NodeHandlerMock struct {
	GetProcessComponentsCalled    func() factory.ProcessComponentsHolder
	GetChainHandlerCalled         func() chainData.ChainHandler
	GetBroadcastMessengerCalled   func() consensus.BroadcastMessenger
	GetShardCoordinatorCalled     func() sharding.Coordinator
	GetCryptoComponentsCalled     func() factory.CryptoComponentsHolder
	GetCoreComponentsCalled       func() factory.CoreComponentsHolder
	GetDataComponentsCalled       func() factory.DataComponentsHandler
	GetStateComponentsCalled      func() factory.StateComponentsHolder
	GetFacadeHandlerCalled        func() shared.FacadeHandler
	GetStatusCoreComponentsCalled func() factory.StatusCoreComponentsHolder
	SetKeyValueForAddressCalled   func(addressBytes []byte, state map[string]string) error
	SetStateForAddressCalled      func(address []byte, state *dtos.AddressState) error
	CloseCalled                   func() error
}

// GetProcessComponents -
func (mock *NodeHandlerMock) GetProcessComponents() factory.ProcessComponentsHolder {
	if mock.GetProcessComponentsCalled != nil {
		return mock.GetProcessComponentsCalled()
	}
	return nil
}

// GetChainHandler -
func (mock *NodeHandlerMock) GetChainHandler() chainData.ChainHandler {
	if mock.GetChainHandlerCalled != nil {
		return mock.GetChainHandlerCalled()
	}
	return nil
}

// GetBroadcastMessenger -
func (mock *NodeHandlerMock) GetBroadcastMessenger() consensus.BroadcastMessenger {
	if mock.GetBroadcastMessengerCalled != nil {
		return mock.GetBroadcastMessengerCalled()
	}
	return nil
}

// GetShardCoordinator -
func (mock *NodeHandlerMock) GetShardCoordinator() sharding.Coordinator {
	if mock.GetShardCoordinatorCalled != nil {
		return mock.GetShardCoordinatorCalled()
	}
	return nil
}

// GetCryptoComponents -
func (mock *NodeHandlerMock) GetCryptoComponents() factory.CryptoComponentsHolder {
	if mock.GetCryptoComponentsCalled != nil {
		return mock.GetCryptoComponentsCalled()
	}
	return nil
}

// GetCoreComponents -
func (mock *NodeHandlerMock) GetCoreComponents() factory.CoreComponentsHolder {
	if mock.GetCoreComponentsCalled != nil {
		return mock.GetCoreComponentsCalled()
	}
	return nil
}

// GetDataComponents -
func (mock *NodeHandlerMock) GetDataComponents() factory.DataComponentsHolder {
	if mock.GetDataComponentsCalled != nil {
		return mock.GetDataComponentsCalled()
	}
	return nil
}

// GetStateComponents -
func (mock *NodeHandlerMock) GetStateComponents() factory.StateComponentsHolder {
	if mock.GetStateComponentsCalled != nil {
		return mock.GetStateComponentsCalled()
	}
	return nil
}

// GetFacadeHandler -
func (mock *NodeHandlerMock) GetFacadeHandler() shared.FacadeHandler {
	if mock.GetFacadeHandlerCalled != nil {
		return mock.GetFacadeHandlerCalled()
	}
	return nil
}

// GetStatusCoreComponents -
func (mock *NodeHandlerMock) GetStatusCoreComponents() factory.StatusCoreComponentsHolder {
	if mock.GetStatusCoreComponentsCalled != nil {
		return mock.GetStatusCoreComponentsCalled()
	}
	return nil
}

// SetKeyValueForAddress -
func (mock *NodeHandlerMock) SetKeyValueForAddress(addressBytes []byte, state map[string]string) error {
	if mock.SetKeyValueForAddressCalled != nil {
		return mock.SetKeyValueForAddressCalled(addressBytes, state)
	}
	return nil
}

// SetStateForAddress -
func (mock *NodeHandlerMock) SetStateForAddress(address []byte, state *dtos.AddressState) error {
	if mock.SetStateForAddressCalled != nil {
		return mock.SetStateForAddressCalled(address, state)
	}
	return nil
}

// Close -
func (mock *NodeHandlerMock) Close() error {
	if mock.CloseCalled != nil {
		return mock.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (mock *NodeHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
