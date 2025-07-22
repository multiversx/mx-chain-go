package txcache

import "github.com/multiversx/mx-chain-go/state"

// breadcrumbsValidator checks that breadcrumbs are continuous
//
//	with the session nonce
//	with the previous breadcrumbs
type breadcrumbsValidator struct {
	skippedSenders                      map[string]struct{}
	sendersInContinuityWithSessionNonce map[string]struct{}
	accountPreviousBreadcrumb           map[string]*accountBreadcrumb
}

func newBreadcrumbValidator() *breadcrumbsValidator {
	return &breadcrumbsValidator{
		skippedSenders:                      make(map[string]struct{}),
		sendersInContinuityWithSessionNonce: make(map[string]struct{}),
		accountPreviousBreadcrumb:           make(map[string]*accountBreadcrumb),
	}
}

// continuousBreadcrumb is used when a block is proposed and also when the deriveVirtualSession is called
func (validator *breadcrumbsValidator) continuousBreadcrumb(
	address string,
	breadcrumb *accountBreadcrumb,
	accountState state.UserAccountHandler,
) bool {

	if breadcrumb.hasUnkownNonce() {
		return true
	}

	accountNonce := accountState.GetNonce()

	if !validator.continuousWithSessionNonce(address, accountNonce, breadcrumb) {
		return false
	}

	if !validator.continuousWithPreviousBreadcrumb(address, accountNonce, breadcrumb) {
		return false
	}

	return true
}

func (validator *breadcrumbsValidator) continuousWithSessionNonce(
	address string,
	accountNonce uint64,
	breadcrumb *accountBreadcrumb,
) bool {

	_, ok := validator.sendersInContinuityWithSessionNonce[address]
	continuousWithSessionNonce := breadcrumb.verifyContinuityWithSessionNonce(accountNonce)
	if !ok && !continuousWithSessionNonce {
		validator.skippedSenders[address] = struct{}{}
		log.Debug("virtualSessionProvider.continuousBreadcrumb breadcrumb not continuous with session nonce",
			"address", address,
			"accountNonce", accountNonce,
			"breadcrumb nonce", breadcrumb.initialNonce)
		return false
	}
	if !ok {
		validator.sendersInContinuityWithSessionNonce[address] = struct{}{}
	}

	return true
}

func (validator *breadcrumbsValidator) continuousWithPreviousBreadcrumb(
	address string,
	accountNonce uint64,
	breadcrumb *accountBreadcrumb) bool {

	previousBreadcrumb := validator.accountPreviousBreadcrumb[address]
	continuousBreadcrumbs := breadcrumb.verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumb)
	if !continuousBreadcrumbs {
		validator.skippedSenders[address] = struct{}{}
		log.Debug("virtualSessionProvider.continuousBreadcrumb breadcrumb not continuous with previous breadcrumb",
			"address", address,
			"accountNonce", accountNonce,
			"current breadcrumb nonce", breadcrumb.initialNonce,
			"previous breadcrumb nonce", previousBreadcrumb.initialNonce)
		return false
	}

	validator.accountPreviousBreadcrumb[address] = breadcrumb

	return true
}
