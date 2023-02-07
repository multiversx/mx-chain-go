package nodesCoordinator

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type nodesCoordinatorRegistryFactory struct {
	marshaller                marshal.Marshalizer
	stakingV4Step2EnableEpoch uint32
}

// NewNodesCoordinatorRegistryFactory creates a nodes coordinator registry factory which will create a
// NodesCoordinatorRegistryHandler from a buffer depending on the epoch
func NewNodesCoordinatorRegistryFactory(
	marshaller marshal.Marshalizer,
	stakingV4Step2EnableEpoch uint32,
) (*nodesCoordinatorRegistryFactory, error) {
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshalizer
	}

	return &nodesCoordinatorRegistryFactory{
		marshaller:                marshaller,
		stakingV4Step2EnableEpoch: stakingV4Step2EnableEpoch,
	}, nil
}

// CreateNodesCoordinatorRegistry creates a NodesCoordinatorRegistryHandler depending on the buffer. Old version uses
// NodesCoordinatorRegistry with a json marshaller; while the new version(from staking v4) uses NodesCoordinatorRegistryWithAuction
// with proto marshaller
func (ncf *nodesCoordinatorRegistryFactory) CreateNodesCoordinatorRegistry(buff []byte) (NodesCoordinatorRegistryHandler, error) {
	registry, err := ncf.createRegistryWithAuction(buff)
	if err == nil {
		log.Debug("nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry created registry with auction",
			"epoch", registry.CurrentEpoch)
		return registry, nil
	}
	log.Debug("nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry creating old registry")
	return createOldRegistry(buff)
}

func (ncf *nodesCoordinatorRegistryFactory) createRegistryWithAuction(buff []byte) (*NodesCoordinatorRegistryWithAuction, error) {
	registry := &NodesCoordinatorRegistryWithAuction{}
	err := ncf.marshaller.Unmarshal(registry, buff)
	if err != nil {
		return nil, err
	}

	log.Debug("nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry created old registry",
		"epoch", registry.CurrentEpoch)
	return registry, nil
}

func createOldRegistry(buff []byte) (*NodesCoordinatorRegistry, error) {
	registry := &NodesCoordinatorRegistry{}
	err := json.Unmarshal(buff, registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

// GetRegistryData returns the registry data as buffer. Old version uses json marshaller, while new version uses proto marshaller
func (ncf *nodesCoordinatorRegistryFactory) GetRegistryData(registry NodesCoordinatorRegistryHandler, epoch uint32) ([]byte, error) {
	if epoch >= ncf.stakingV4Step2EnableEpoch {
		log.Debug("nodesCoordinatorRegistryFactory.GetRegistryData called with auction after staking v4", "epoch", epoch)
		return ncf.marshaller.Marshal(registry)
	}
	log.Debug("nodesCoordinatorRegistryFactory.GetRegistryData called with old json before staking v4", "epoch", epoch)
	return json.Marshal(registry)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ncf *nodesCoordinatorRegistryFactory) IsInterfaceNil() bool {
	return ncf == nil
}
