package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state/syncer"
)

// ValidatorAccountsSyncerFactoryMock -
type ValidatorAccountsSyncerFactoryMock struct {
	CreateValidatorAccountsSyncerCalled func(args syncer.ArgsNewValidatorAccountsSyncer) (process.AccountsDBSyncer, error)
}

// CreateValidatorAccountsSyncer -
func (mock *ValidatorAccountsSyncerFactoryMock) CreateValidatorAccountsSyncer(args syncer.ArgsNewValidatorAccountsSyncer) (process.AccountsDBSyncer, error) {
	if mock.CreateValidatorAccountsSyncerCalled != nil {
		return mock.CreateValidatorAccountsSyncerCalled(args)
	}
	return &processMock.AccountsDBSyncerStub{}, nil
}

// IsInterfaceNil -
func (mock *ValidatorAccountsSyncerFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
