package txcache

import (
	"math/big"
)

type virtualSessionProvider struct {
	session                  SelectionSession
	validator                *breadcrumbsValidator
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

func newVirtualSessionProvider(session SelectionSession) *virtualSessionProvider {
	return &virtualSessionProvider{
		session:                  session,
		validator:                newBreadcrumbValidator(),
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

func (provider *virtualSessionProvider) createVirtualSelectionSession(
	chainOfTrackedBlocks []*trackedBlock,
) (*virtualSelectionSession, error) {

	for _, tb := range chainOfTrackedBlocks {
		err := provider.handleTrackedBlock(tb)
		if err != nil {
			return nil, err
		}
	}

	virtualSession := newVirtualSelectionSession(provider.session)
	virtualSession.virtualAccountsByAddress = provider.virtualAccountsByAddress
	return virtualSession, nil
}

func (provider *virtualSessionProvider) handleTrackedBlock(tb *trackedBlock) error {
	for address, breadcrumb := range tb.breadcrumbsByAddress {
		ok := provider.validator.shouldSkipSender(address)
		if ok {
			continue
		}

		accountNonce, accountBalance, _, err := provider.session.GetAccountNonceAndBalance([]byte(address))
		if err != nil {
			log.Debug("virtualSessionProvider.handleTrackedBlock",
				"err", err,
				"address", address,
				"tracked block rootHash", tb.rootHash)
			return err
		}

		if !provider.validator.continuousBreadcrumb(address, accountNonce, breadcrumb) {
			delete(provider.virtualAccountsByAddress, address)
			continue
		}

		provider.fromBreadcrumbToVirtualRecord(address, accountBalance, breadcrumb)
	}

	return nil
}

func (provider *virtualSessionProvider) fromBreadcrumbToVirtualRecord(
	address string,
	accountBalance *big.Int,
	breadcrumb *accountBreadcrumb,
) {
	virtualRecord, ok := provider.virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountBalance
		virtualRecord = newVirtualAccountRecord(breadcrumb.initialNonce, initialBalance)
		provider.virtualAccountsByAddress[address] = virtualRecord
	}

	virtualRecord.updateVirtualRecord(breadcrumb)
}
