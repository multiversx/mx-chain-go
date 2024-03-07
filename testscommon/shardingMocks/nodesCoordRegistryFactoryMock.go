package shardingMocks

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// NodesCoordinatorRegistryFactoryMock -
type NodesCoordinatorRegistryFactoryMock struct {
}

// CreateNodesCoordinatorRegistry -
func (ncr *NodesCoordinatorRegistryFactoryMock) CreateNodesCoordinatorRegistry(buff []byte) (nodesCoordinator.NodesCoordinatorRegistryHandler, error) {
	registry := &nodesCoordinator.NodesCoordinatorRegistry{}
	err := json.Unmarshal(buff, registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

// GetRegistryData -
func (ncr *NodesCoordinatorRegistryFactoryMock) GetRegistryData(registry nodesCoordinator.NodesCoordinatorRegistryHandler, _ uint32) ([]byte, error) {
	return json.Marshal(registry)
}

// IsInterfaceNil -
func (ncr *NodesCoordinatorRegistryFactoryMock) IsInterfaceNil() bool {
	return ncr == nil
}
