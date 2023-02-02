package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// BootstrapComponentsStub -
type BootstrapComponentsStub struct {
	Bootstrapper               factory.EpochStartBootstrapper
	BootstrapParams            factory.BootstrapParamsHolder
	NodeRole                   core.NodeType
	ShCoordinator              sharding.Coordinator
	HdrVersionHandler          nodeFactory.HeaderVersionHandler
	VersionedHdrFactory        nodeFactory.VersionedHeaderFactory
	HdrIntegrityVerifier       nodeFactory.HeaderIntegrityVerifierHandler
	GuardedAccountHandlerField process.GuardedAccountHandler
}

// Create -
func (bcs *BootstrapComponentsStub) Create() error {
	return nil
}

// Close -
func (bcs *BootstrapComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (bcs *BootstrapComponentsStub) CheckSubcomponents() error {
	return nil
}

// EpochStartBootstrapper -
func (bcs *BootstrapComponentsStub) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	return bcs.Bootstrapper
}

// EpochBootstrapParams -
func (bcs *BootstrapComponentsStub) EpochBootstrapParams() factory.BootstrapParamsHolder {
	return bcs.BootstrapParams
}

// NodeType -
func (bcs *BootstrapComponentsStub) NodeType() core.NodeType {
	return bcs.NodeRole
}

// ShardCoordinator -
func (bcs *BootstrapComponentsStub) ShardCoordinator() sharding.Coordinator {
	return bcs.ShCoordinator
}

// HeaderVersionHandler -
func (bcs *BootstrapComponentsStub) HeaderVersionHandler() nodeFactory.HeaderVersionHandler {
	return bcs.HdrVersionHandler
}

// VersionedHeaderFactory -
func (bcs *BootstrapComponentsStub) VersionedHeaderFactory() nodeFactory.VersionedHeaderFactory {
	return bcs.VersionedHdrFactory
}

// HeaderIntegrityVerifier -
func (bcs *BootstrapComponentsStub) HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler {
	return bcs.HdrIntegrityVerifier
}

// SetShardCoordinator -
func (bcs *BootstrapComponentsStub) SetShardCoordinator(shardCoordinator sharding.Coordinator) error {
	bcs.ShCoordinator = shardCoordinator
	return nil
}

// GuardedAccountHandler -
func (bcs *BootstrapComponentsStub) GuardedAccountHandler() process.GuardedAccountHandler {
	return bcs.GuardedAccountHandlerField
}

// String -
func (bcs *BootstrapComponentsStub) String() string {
	return "BootstrapComponentsStub"
}

// IsInterfaceNil -
func (bcs *BootstrapComponentsStub) IsInterfaceNil() bool {
	return bcs == nil
}
