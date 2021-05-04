package mainFactoryMocks

import (
	nodeFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BootstrapComponentsStub -
type BootstrapComponentsStub struct {
	Bootstrapper         factory.EpochStartBootstrapper
	BootstrapParams      factory.BootstrapParamsHolder
	NodeRole             core.NodeType
	ShCoordinator        sharding.Coordinator
	HdrIntegrityVerifier nodeFactory.HeaderIntegrityVerifierHandler
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

// HeaderIntegrityVerifier -
func (bcs *BootstrapComponentsStub) HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler {
	return bcs.HdrIntegrityVerifier
}

// IsInterfaceNil -
func (bcs *BootstrapComponentsStub) IsInterfaceNil() bool {
	return bcs == nil
}
