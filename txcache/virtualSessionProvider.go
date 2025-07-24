package txcache

import "github.com/multiversx/mx-chain-go/state"

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

		// TODO make sure that the accounts which don't yet exist are properly handled
		accountState, err := provider.session.GetAccountState([]byte(address))
		if err != nil {
			log.Debug("virtualSessionProvider.handleTrackedBlock",
				"err", err)
			return err
		}

		if !provider.validator.continuousBreadcrumb(address, breadcrumb, accountState) {
			delete(provider.virtualAccountsByAddress, address)
			continue
		}

		provider.handleAccountBreadcrumb(breadcrumb, accountState, address)
	}

	return nil
}

func (provider *virtualSessionProvider) handleAccountBreadcrumb(
	breadcrumb *accountBreadcrumb,
	accountState state.UserAccountHandler,
	address string,
) {
	virtualRecord, ok := provider.virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountState.GetBalance()
		virtualRecord = newVirtualAccountRecord(breadcrumb.initialNonce, initialBalance)
		provider.virtualAccountsByAddress[address] = virtualRecord
	}

	virtualRecord.updateVirtualRecord(breadcrumb)
}
