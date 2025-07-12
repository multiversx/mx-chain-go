package txcache

import "github.com/multiversx/mx-chain-go/state"

type virtualSessionProvider struct {
	session                             SelectionSession
	virtualAccountsByAddress            map[string]*virtualAccountRecord
	skippedSenders                      map[string]struct{}
	sendersInContinuityWithSessionNonce map[string]struct{}
	accountPreviousBreadcrumb           map[string]*accountBreadcrumb
}

func newVirtualSessionProvider(session SelectionSession) *virtualSessionProvider {
	return &virtualSessionProvider{
		session:                             session,
		virtualAccountsByAddress:            make(map[string]*virtualAccountRecord),
		skippedSenders:                      make(map[string]struct{}),
		sendersInContinuityWithSessionNonce: make(map[string]struct{}),
		accountPreviousBreadcrumb:           make(map[string]*accountBreadcrumb),
	}
}

func (provider *virtualSessionProvider) handleTrackedBlock(tb *trackedBlock) error {
	for address, breadcrumb := range tb.breadcrumbsByAddress {
		_, ok := provider.skippedSenders[address]
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

		accountNonce := accountState.GetNonce()

		if !provider.continuousBreadcrumb(address, breadcrumb, accountNonce) {
			provider.skippedSenders[address] = struct{}{}
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

func (provider *virtualSessionProvider) continuousBreadcrumb(
	address string,
	breadcrumb *accountBreadcrumb,
	accountNonce uint64,
) bool {
	if breadcrumb.hasUnkownNonce() {
		return true
	}

	_, ok := provider.sendersInContinuityWithSessionNonce[address]
	continuousWithSessionNonce := breadcrumb.verifyContinuityWithSessionNonce(accountNonce)
	if !ok && !continuousWithSessionNonce {
		log.Debug("virtualSessionProvider.continuousBreadcrumb breadcrumb not continuous with session nonce",
			"address", address,
			"accountNonce", accountNonce,
			"breadcrumb nonce", breadcrumb.initialNonce)
		return false
	}
	if !ok {
		provider.sendersInContinuityWithSessionNonce[address] = struct{}{}
	}

	previousBreadcrumb := provider.accountPreviousBreadcrumb[address]
	continuousBreadcrumbs := breadcrumb.verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumb)
	if !continuousBreadcrumbs {
		log.Debug("virtualSessionProvider.continuousBreadcrumb breadcrumb not continuous with previous breadcrumb",
			"address", address,
			"accountNonce", accountNonce,
			"current breadcrumb nonce", breadcrumb.initialNonce,
			"previous breadcrumb nonce", previousBreadcrumb.initialNonce)
		return false
	}

	provider.accountPreviousBreadcrumb[address] = breadcrumb
	return true
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
