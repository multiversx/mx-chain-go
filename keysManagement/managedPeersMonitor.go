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
}

type managedPeersMonitor struct {
	managedPeersHolder common.ManagedPeersHolder
	nodesCoordinator   NodesCoordinator
	shardProvider      ShardProvider
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

	return nil
}

// GetManagedKeysCount returns the number of keys managed by the current node
func (monitor *managedPeersMonitor) GetManagedKeysCount() int {
	return len(monitor.managedPeersHolder.GetManagedKeysByCurrentNode())
}

// GetEligibleManagedKeys returns eligible keys that are managed by the current node
func (monitor *managedPeersMonitor) GetEligibleManagedKeys(epoch uint32) ([][]byte, error) {
	eligibleValidators, err := monitor.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	return monitor.extractManagedIntraShardKeys(eligibleValidators)
}

// GetWaitingManagedKeys returns waiting keys that are managed by the current node
func (monitor *managedPeersMonitor) GetWaitingManagedKeys(epoch uint32) ([][]byte, error) {
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
