package mainFactoryMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	nodeFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BootstrapComponentsStub -
type BootstrapComponentsStub struct {
	Bootstrapper                factory.EpochStartBootstrapper
	BootstrapParams             factory.BootstrapParamsHolder
	NodeRole                    core.NodeType
	ShCoordinator               sharding.Coordinator
	HdrVersionHandler           nodeFactory.HeaderVersionHandler
	VersionedHdrFactory         nodeFactory.VersionedHeaderFactory
	HdrIntegrityVerifier        nodeFactory.HeaderIntegrityVerifierHandler
	RoundActivationHandlerField process.RoundActivationHandler
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
func (bc *BootstrapComponentsStub) VersionedHeaderFactory() nodeFactory.VersionedHeaderFactory {
	return bc.VersionedHdrFactory
}

// HeaderIntegrityVerifier -
func (bcs *BootstrapComponentsStub) HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler {
	return bcs.HdrIntegrityVerifier
}

// RoundActivationHandler -
func (bcs *BootstrapComponentsStub) RoundActivationHandler() process.RoundActivationHandler {
	return bcs.RoundActivationHandlerField
}

// String -
func (bcs *BootstrapComponentsStub) String() string {
	return "BootstrapComponentsStub"
}

// IsInterfaceNil -
func (bcs *BootstrapComponentsStub) IsInterfaceNil() bool {
	return bcs == nil
}
