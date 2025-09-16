package txcache

import (
	"math/big"
)

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

		if !computer.validator.isContinuousBreadcrumb(address, accountNonce, breadcrumb) {
			delete(computer.virtualAccountsByAddress, address)
			continue
		}

		computer.fromBreadcrumbToVirtualRecord(address, accountBalance, breadcrumb)
	}

	return nil
}

func (computer *virtualSessionComputer) fromBreadcrumbToVirtualRecord(
	address string,
	accountBalance *big.Int,
	breadcrumb *accountBreadcrumb,
) {
	virtualRecord, ok := computer.virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountBalance
		virtualRecord = newVirtualAccountRecord(breadcrumb.initialNonce, initialBalance)
		computer.virtualAccountsByAddress[address] = virtualRecord
	}

	virtualRecord.updateVirtualRecord(breadcrumb)
}
