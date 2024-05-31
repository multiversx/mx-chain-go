package keysManagement

import (
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

// ArgManagedPeersMonitor is the DTO used to create a new instance of managedPeersMonitor
type ArgManagedPeersMonitor struct {
	ManagedPeersHolder common.ManagedPeersHolder
	NodesCoordinator   NodesCoordinator
	ShardProvider      ShardProvider
	EpochProvider      CurrentEpochProvider
}

type managedPeersMonitor struct {
	managedPeersHolder common.ManagedPeersHolder
	nodesCoordinator   NodesCoordinator
	shardProvider      ShardProvider
	epochProvider      CurrentEpochProvider
}

// NewManagedPeersMonitor returns a new instance of managedPeersMonitor
func NewManagedPeersMonitor(args ArgManagedPeersMonitor) (*managedPeersMonitor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &managedPeersMonitor{
		managedPeersHolder: args.ManagedPeersHolder,
		nodesCoordinator:   args.NodesCoordinator,
		shardProvider:      args.ShardProvider,
		epochProvider:      args.EpochProvider,
	}, nil
}

func checkArgs(args ArgManagedPeersMonitor) error {
	if check.IfNil(args.ManagedPeersHolder) {
		return ErrNilManagedPeersHolder
	}
	if check.IfNil(args.NodesCoordinator) {
		return ErrNilNodesCoordinator
	}
	if check.IfNil(args.ShardProvider) {
		return ErrNilShardProvider
	}
	if check.IfNil(args.EpochProvider) {
		return ErrNilEpochProvider
	}

	return nil
}

// GetManagedKeysCount returns the number of keys managed by the current node
func (monitor *managedPeersMonitor) GetManagedKeysCount() int {
	return len(monitor.managedPeersHolder.GetManagedKeysByCurrentNode())
}

// GetManagedKeys returns all keys that should act as validator(main or backup that took over) and will be managed by this node
func (monitor *managedPeersMonitor) GetManagedKeys() [][]byte {
	managedKeysMap := monitor.managedPeersHolder.GetManagedKeysByCurrentNode()
	managedKeys := make([][]byte, 0, len(managedKeysMap))
	for pk := range managedKeysMap {
		managedKeys = append(managedKeys, []byte(pk))
	}

	sort.Slice(managedKeys, func(i, j int) bool {
		return string(managedKeys[i]) < string(managedKeys[j])
	})

	return managedKeys
}

// GetLoadedKeys returns all keys that were loaded and will be managed by this node
func (monitor *managedPeersMonitor) GetLoadedKeys() [][]byte {
	return monitor.managedPeersHolder.GetLoadedKeysByCurrentNode()
}

// GetEligibleManagedKeys returns eligible keys that are managed by the current node in the current epoch
func (monitor *managedPeersMonitor) GetEligibleManagedKeys() ([][]byte, error) {
	epoch := monitor.epochProvider.CurrentEpoch()
	eligibleValidators, err := monitor.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	return monitor.extractManagedIntraShardKeys(eligibleValidators)
}

// GetWaitingManagedKeys returns waiting keys that are managed by the current node
func (monitor *managedPeersMonitor) GetWaitingManagedKeys() ([][]byte, error) {
	epoch := monitor.epochProvider.CurrentEpoch()
	waitingValidators, err := monitor.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	return monitor.extractManagedIntraShardKeys(waitingValidators)
}

func (monitor *managedPeersMonitor) extractManagedIntraShardKeys(keysMap map[uint32][][]byte) ([][]byte, error) {
	selfShardID := monitor.shardProvider.SelfId()
	intraShardKeys, ok := keysMap[selfShardID]
	if !ok {
		return nil, fmt.Errorf("%w for shard %d, no validators found", ErrInvalidValue, selfShardID)
	}

	managedKeys := make([][]byte, 0)
	for _, key := range intraShardKeys {
		if monitor.managedPeersHolder.IsKeyManagedByCurrentNode(key) {
			managedKeys = append(managedKeys, key)
		}
	}

	sort.Slice(managedKeys, func(i, j int) bool {
		return string(managedKeys[i]) < string(managedKeys[j])
	})

	return managedKeys, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *managedPeersMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
