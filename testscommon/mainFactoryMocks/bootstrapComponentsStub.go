package mainFactoryMocks

import (
	"github.com/ElrondNetwork/elrond-go/core"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BootstrapComponentsStub -
type BootstrapComponentsStub struct {
	Bootstrapper         mainFactory.EpochStartBootstrapper
	BootstrapParams      mainFactory.BootstrapParamsHandler
	NodeRole             core.NodeType
	ShCoordinator        sharding.Coordinator
	HdrIntegrityVerifier mainFactory.HeaderIntegrityVerifierHandler
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
func (bcs *BootstrapComponentsStub) EpochStartBootstrapper() mainFactory.EpochStartBootstrapper {
	return bcs.Bootstrapper
}

// EpochBootstrapParams -
func (bcs *BootstrapComponentsStub) EpochBootstrapParams() mainFactory.BootstrapParamsHandler {
	return bcs.BootstrapParams
}

// NodeType -
func (bcs *BootstrapComponentsStub) NodeType() core.NodeType{
	return bcs.NodeRole
}

// ShardCoordinator -
func (bcs *BootstrapComponentsStub) ShardCoordinator() sharding.Coordinator{
	return bcs.ShCoordinator
}

// HeaderIntegrityVerifier -
func (bcs *BootstrapComponentsStub) HeaderIntegrityVerifier() mainFactory.HeaderIntegrityVerifierHandler{
	return bcs.HdrIntegrityVerifier
}

// IsInterfaceNil -
func (bcs *BootstrapComponentsStub) IsInterfaceNil() bool {
	return bcs == nil
}
