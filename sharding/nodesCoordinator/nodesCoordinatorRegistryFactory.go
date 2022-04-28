package nodesCoordinator

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type nodesCoordinatorRegistryFactory struct {
	stakingV4EnableEpoch uint32
	flagStakingV4        atomic.Flag
	marshaller           marshal.Marshalizer
}

// NewNodesCoordinatorRegistryFactory creates a nodes coordinator registry factory which will create a
// NodesCoordinatorRegistryHandler from a buffer depending on the epoch
func NewNodesCoordinatorRegistryFactory(
	marshaller marshal.Marshalizer,
	notifier EpochNotifier,
	stakingV4EnableEpoch uint32,
) (*nodesCoordinatorRegistryFactory, error) {
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(notifier) {
		return nil, ErrNilEpochNotifier
	}

	log.Debug("nodesCoordinatorRegistryFactory: staking v4 enable epoch", "epoch", stakingV4EnableEpoch)

	ncf := &nodesCoordinatorRegistryFactory{
		marshaller:           marshaller,
		stakingV4EnableEpoch: stakingV4EnableEpoch,
	}
	notifier.RegisterNotifyHandler(ncf)
	return ncf, nil
}

// CreateNodesCoordinatorRegistry creates a NodesCoordinatorRegistryHandler depending on the buffer. Old version uses
// NodesCoordinatorRegistry with a json marshaller; while the new version(from staking v4) uses NodesCoordinatorRegistryWithAuction
// with proto marshaller
func (ncf *nodesCoordinatorRegistryFactory) CreateNodesCoordinatorRegistry(buff []byte) (NodesCoordinatorRegistryHandler, error) {
	//if ncf.flagStakingV4.IsSet() {
	//	return ncf.createRegistryWithAuction(buff)
	//}
	//return createOldRegistry(buff)
	registry, err := ncf.createRegistryWithAuction(buff)
	if err == nil {
		log.Debug("nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry created registry with auction")
		return registry, nil
	}
	log.Debug("nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry created old registry")
	return createOldRegistry(buff)
}

func (ncf *nodesCoordinatorRegistryFactory) GetRegistryData(registry NodesCoordinatorRegistryHandler, epoch uint32) ([]byte, error) {
	if epoch >= ncf.stakingV4EnableEpoch {
		log.Debug("nodesCoordinatorRegistryFactory.GetRegistryData called with auction", "epoch", epoch)
		return ncf.marshaller.Marshal(registry)
	}
	log.Debug("nodesCoordinatorRegistryFactory.GetRegistryData called with old json", "epoch", epoch)
	return json.Marshal(registry)
}

func createOldRegistry(buff []byte) (*NodesCoordinatorRegistry, error) {
	registry := &NodesCoordinatorRegistry{}
	err := json.Unmarshal(buff, registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (ncf *nodesCoordinatorRegistryFactory) createRegistryWithAuction(buff []byte) (*NodesCoordinatorRegistryWithAuction, error) {
	registry := &NodesCoordinatorRegistryWithAuction{}
	err := ncf.marshaller.Unmarshal(registry, buff)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ncf *nodesCoordinatorRegistryFactory) IsInterfaceNil() bool {
	return ncf == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ncf *nodesCoordinatorRegistryFactory) EpochConfirmed(epoch uint32, _ uint64) {
	ncf.flagStakingV4.SetValue(epoch >= ncf.stakingV4EnableEpoch)
	log.Debug("nodesCoordinatorRegistryFactory: staking v4", "enabled", ncf.flagStakingV4.IsSet())
}
