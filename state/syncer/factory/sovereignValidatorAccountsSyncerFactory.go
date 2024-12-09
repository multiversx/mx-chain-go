package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state/syncer"
)

type sovereignValidatorAccountsSyncerFactory struct {
}

// NewSovereignValidatorAccountsSyncerFactory creates a new sovereign validator accounts syncer factory
func NewSovereignValidatorAccountsSyncerFactory() *sovereignValidatorAccountsSyncerFactory {
	return &sovereignValidatorAccountsSyncerFactory{}
}

// CreateValidatorAccountsSyncer creates a sovereign validator accounts syncer
func (f *sovereignValidatorAccountsSyncerFactory) CreateValidatorAccountsSyncer(args syncer.ArgsNewValidatorAccountsSyncer) (process.AccountsDBSyncer, error) {
	return syncer.NewSovereignValidatorAccountsSyncer(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignValidatorAccountsSyncerFactory) IsInterfaceNil() bool {
	return f == nil
}
