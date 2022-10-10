package shared

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/state"
)

// errInvalidAlteredAccountsOptions signals that invalid options has been passed when extracting altered accounts
var errInvalidAlteredAccountsOptions = errors.New("invalid altered accounts options")

// AlteredAccountsOptions holds some configurable parameters to be used for extracting the altered accounts
type AlteredAccountsOptions struct {
	WithCustomAccountsRepository bool
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
