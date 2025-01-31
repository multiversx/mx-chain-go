package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state/syncer"
)

type validatorAccountsSyncerFactory struct {
}

// NewValidatorAccountsSyncerFactory creates a new validator accounts syncer factory
func NewValidatorAccountsSyncerFactory() *validatorAccountsSyncerFactory {
	return &validatorAccountsSyncerFactory{}
}

// CreateValidatorAccountsSyncer creates a validator accounts syncer
func (f *validatorAccountsSyncerFactory) CreateValidatorAccountsSyncer(args syncer.ArgsNewValidatorAccountsSyncer) (process.AccountsDBSyncer, error) {
	return syncer.NewValidatorAccountsSyncer(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *validatorAccountsSyncerFactory) IsInterfaceNil() bool {
	return f == nil
}
