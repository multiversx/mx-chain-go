package txcache

import (
	"math/big"
)

// virtualSessionComputer relies on the internal state of the validator for skipping certain senders
type virtualSessionComputer struct {
	session                  SelectionSession
	validator                *breadcrumbsValidator
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

func newVirtualSessionComputer(session SelectionSession) *virtualSessionComputer {
	return &virtualSessionComputer{
		session:                  session,
		validator:                newBreadcrumbValidator(),
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

// createVirtualSelectionSession iterates over the chain of tracked blocks and for each sender checks that its breadcrumb is continuous
// If the breadcrumb of an account is continuous, the virtual record of that account is created or updated
func (computer *virtualSessionComputer) createVirtualSelectionSession(
	chainOfTrackedBlocks []*trackedBlock,
) (*virtualSelectionSession, error) {
	for _, tb := range chainOfTrackedBlocks {
		err := computer.handleTrackedBlock(tb)
		if err != nil {
			return nil, err
		}
	}

	virtualSession := newVirtualSelectionSession(computer.session)
	virtualSession.virtualAccountsByAddress = computer.virtualAccountsByAddress
	return virtualSession, nil
}

func (computer *virtualSessionComputer) handleTrackedBlock(tb *trackedBlock) error {
	for address, breadcrumb := range tb.breadcrumbsByAddress {
		// check if this address was already marked as not continuous by the validator
		ok := computer.validator.shouldSkipSender(address)
		if ok {
			continue
		}

		accountNonce, accountBalance, _, err := computer.session.GetAccountNonceAndBalance([]byte(address))
		if err != nil {
			log.Debug("virtualSessionComputer.handleTrackedBlock",
				"err", err,
				"address", address,
				"tracked block rootHash", tb.rootHash)
			return err
		}

		if !computer.validator.validateNonceContinuityOfBreadcrumb(address, accountNonce, breadcrumb) {
			delete(computer.virtualAccountsByAddress, address)
			continue
		}

		err = computer.fromBreadcrumbToVirtualRecord(address, accountBalance, breadcrumb)
		if err != nil {
			return err
		}
	}

	return nil
}

func (computer *virtualSessionComputer) fromBreadcrumbToVirtualRecord(
	address string,
	accountBalance *big.Int,
	breadcrumb *accountBreadcrumb,
) error {
	virtualRecord, ok := computer.virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountBalance
		record, err := newVirtualAccountRecord(breadcrumb.initialNonce, initialBalance)
		if err != nil {
			return err
		}

		virtualRecord = record
		computer.virtualAccountsByAddress[address] = record
	}

	virtualRecord.updateVirtualRecord(breadcrumb)
	return nil
}
