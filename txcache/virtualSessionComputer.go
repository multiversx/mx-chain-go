package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

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

// createVirtualSelectionSession iterates over the global breadcrumbs of the selection tracker.
// If the global breadcrumb of an account is continuous with the session nonce,
// the virtual record of that account is created or updated.
// NOTE: The createVirtualSelectionSession method should receive a deep copy of the globalAccountBreadcrumbs.
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

// handleGlobalAccountBreadcrumbs iterates over each global account breadcrumb, verifies the continuity with the session nonce
// and transforms each global account breadcrumb into a virtual record.
func (computer *virtualSessionComputer) handleGlobalAccountBreadcrumbs(
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb,
) error {
	for address, globalBreadcrumb := range globalAccountBreadcrumbs {
		accountNonce, accountBalance, _, err := computer.session.GetAccountNonceAndBalance([]byte(address))
		if err != nil {
			log.Debug("virtualSessionComputer.handleGlobalAccountBreadcrumbs",
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
			continue
		}

		err = computer.fromGlobalBreadcrumbToVirtualRecord(address, accountNonce, accountBalance, globalBreadcrumb)
		if err != nil {
			return err
		}
	}

	return nil
}

// fromGlobalBreadcrumbToVirtualRecord transforms a global account breadcrumb simply by:
// initializing the initialNonce of the virtual record with the latestNonce + 1
// copying the consumed balance in the initialBalance of the virtual record.
func (computer *virtualSessionComputer) fromGlobalBreadcrumbToVirtualRecord(
	address string,
	accountNonce uint64,
	accountBalance *big.Int,
	globalBreadcrumb *globalAccountBreadcrumb,
) error {
	_, ok := computer.virtualAccountsByAddress[address]
	if ok {
		return nil
	}

	initialBalance := accountBalance
	initialNonce := core.OptionalUint64{
		Value:    accountNonce,
		HasValue: true,
	}

	if globalBreadcrumb.isUser() {
		initialNonce = core.OptionalUint64{
			Value:    globalBreadcrumb.lastNonce.Value + 1,
			HasValue: true,
		}
	}

	// We initialize the virtual record with the session nonce because an account might be only a relayer in the proposed blocks.
	// Without this initialization, the initialNonce remains without a value.
	// On the selection side, a virtual account record that has an initial nonce without a value
	// will lead to an incorrect skip of a specific tx where the account is a sender.
	record, err := newVirtualAccountRecord(initialNonce, initialBalance)
	if err != nil {
		log.Debug("virtualSessionComputer.fromGlobalBreadcrumbToVirtualRecord",
			"err", err,
			"address", address,
			"accountBalance", accountBalance,
		)
		return err
	}

	record.accumulateConsumedBalance(globalBreadcrumb.consumedBalance)
	computer.virtualAccountsByAddress[address] = record
	return nil
}
