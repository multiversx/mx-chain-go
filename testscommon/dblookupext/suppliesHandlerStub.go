package dblookupext

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
)

// SuppliesHandlerStub -
type SuppliesHandlerStub struct {
	ProcessLogsCalled   func(blockNonce uint64, logs []*data.LogData) error
	RevertChangesCalled func(header data.HeaderHandler, body data.BodyHandler) error
	GetESDTSupplyCalled func(token string) (*esdtSupply.SupplyESDT, error)
}

// ProcessLogs -
func (stub *SuppliesHandlerStub) ProcessLogs(blockNonce uint64, logs []*data.LogData) error {
	if stub.ProcessLogsCalled != nil {
		return stub.ProcessLogsCalled(blockNonce, logs)
	}

	return nil
}

// RevertChanges -
func (stub *SuppliesHandlerStub) RevertChanges(header data.HeaderHandler, body data.BodyHandler) error {
	if stub.RevertChangesCalled != nil {
		return stub.RevertChangesCalled(header, body)
	}

	return nil
}

// GetESDTSupply -
func (stub *SuppliesHandlerStub) GetESDTSupply(token string) (*esdtSupply.SupplyESDT, error) {
	if stub.GetESDTSupplyCalled != nil {
		return stub.GetESDTSupplyCalled(token)
	}

	return nil, fmt.Errorf("not implemented")
}

// IsInterfaceNil -
func (stub *SuppliesHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
