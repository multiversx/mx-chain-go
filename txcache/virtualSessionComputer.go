package txcache

import (
	"math/big"
)

// virtualSessionComputer relies on the internal state of the validator for skipping certain senders
type virtualSessionComputer struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

func newVirtualSessionComputer(session SelectionSession) *virtualSessionComputer {
	return &virtualSessionComputer{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

// createVirtualSelectionSession iterates over the chain of tracked blocks and for each sender checks that its breadcrumb is continuous
// If the breadcrumb of an account is continuous, the virtual record of that account is created or updated
func (computer *virtualSessionComputer) createVirtualSelectionSession(
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb,
) (*virtualSelectionSession, error) {
	err := computer.handleGlobalAccountBreadcrumbs(globalAccountBreadcrumbs)
	if err != nil {
		return nil, err
	}

	virtualSession := newVirtualSelectionSession(computer.session, computer.virtualAccountsByAddress)
	return virtualSession, nil
}

func (computer *virtualSessionComputer) handleGlobalAccountBreadcrumbs(
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb,
) error {
	for address, globalBreadcrumb := range globalAccountBreadcrumbs {
		accountNonce, accountBalance, _, err := computer.session.GetAccountNonceAndBalance([]byte(address))
		if err != nil {
			log.Debug("virtualSessionComputer.handleTrackedBlock",
				"err", err,
				"address", address)
			return err
		}

		if !globalBreadcrumb.continuousWithSessionNonce(accountNonce) {
			log.Debug("virtualSessionComputer.handleGlobalAccountBreadcrumbs global breadcrumb not continuous with session nonce",
				"address", address,
				"accountNonce", accountNonce,
				"breadcrumb nonce", globalBreadcrumb.firstNonce,
			)
			return errDiscontinuousGlobalBreadcrumbs
		}

		err = computer.fromGlobalBreadcrumbToVirtualRecord(address, accountBalance, globalBreadcrumb)
		if err != nil {
			return err
		}
	}

	return nil
}

func (computer *virtualSessionComputer) fromGlobalBreadcrumbToVirtualRecord(
	address string,
	accountBalance *big.Int,
	globalBreadcrumb *globalAccountBreadcrumb,
) error {
	_, ok := computer.virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountBalance
		record, err := newVirtualAccountRecord(globalBreadcrumb.lastNonce, initialBalance)
		if err != nil {
			return err
		}

		record.updateVirtualRecord(globalBreadcrumb)
		computer.virtualAccountsByAddress[address] = record
	}

	return nil
}
