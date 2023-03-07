package shared

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/state"
)

// errInvalidAlteredAccountsOptions signals that invalid options has been passed when extracting altered accounts
var errInvalidAlteredAccountsOptions = errors.New("invalid altered accounts options")

// AlteredAccountsOptions holds some configurable parameters to be used for extracting the altered accounts
type AlteredAccountsOptions struct {
	WithCustomAccountsRepository bool
	WithAdditionalOutportData    bool
	AccountsRepository           state.AccountsRepository
	AccountQueryOptions          api.AccountQueryOptions
}

// Verify will check the validity of the options
func (o *AlteredAccountsOptions) Verify() error {
	if !o.WithCustomAccountsRepository {
		return nil
	}

	if check.IfNil(o.AccountsRepository) {
		return fmt.Errorf("%w: nil accounts repository", errInvalidAlteredAccountsOptions)
	}

	return nil
}
