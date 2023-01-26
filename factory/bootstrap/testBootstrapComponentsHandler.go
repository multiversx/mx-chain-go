package bootstrap

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/sharding"
)

type testManagedBootstrapComponents struct {
	*managedBootstrapComponents
}

// NewTestManagedBootstrapComponents creates an instance of a test managed bootstrap components
func NewTestManagedBootstrapComponents(bootstrapComponentsFactory *bootstrapComponentsFactory) (*testManagedBootstrapComponents, error) {
	bc, err := NewManagedBootstrapComponents(bootstrapComponentsFactory)
	if err != nil {
		return nil, err
	}
	return &testManagedBootstrapComponents{
		managedBootstrapComponents: bc,
	}, nil
}

// SetShardCoordinator sets the shard coordinator
func (mbf *testManagedBootstrapComponents) SetShardCoordinator(shardCoordinator sharding.Coordinator) error {
	mbf.mutBootstrapComponents.RLock()
	defer mbf.mutBootstrapComponents.RUnlock()

	if mbf.bootstrapComponents == nil {
		return errors.ErrNilBootstrapComponents
	}

	mbf.bootstrapComponents.shardCoordinator = shardCoordinator
	return nil
}
