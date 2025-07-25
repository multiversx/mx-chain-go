package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/state"
)

// breadcrumbsValidator checks that breadcrumbs are continuous
//
//	with the session nonce
//	with the previous breadcrumbs
type breadcrumbsValidator struct {
	skippedSenders                      map[string]struct{}
	sendersInContinuityWithSessionNonce map[string]struct{}
	accountPreviousBreadcrumb           map[string]*accountBreadcrumb
	virtualBalancesByAddress            map[string]*virtualAccountBalance
}

// a validator object is used for the next scenarios:
//
// when the creation of the virtual session - deriveVirtualSelectionSession - is called (in the SelectTransactions)
// when the validation for a proposed block is called (when receiving the OnProposedBlock notification)
// at the end of those methods it becomes useless
func newBreadcrumbValidator() *breadcrumbsValidator {
	return &breadcrumbsValidator{
		skippedSenders:                      make(map[string]struct{}),
		sendersInContinuityWithSessionNonce: make(map[string]struct{}),
		accountPreviousBreadcrumb:           make(map[string]*accountBreadcrumb),
		virtualBalancesByAddress:            make(map[string]*virtualAccountBalance),
	}
}

// continuousBreadcrumb is used when a block is proposed and also when the deriveVirtualSession is called
func (validator *breadcrumbsValidator) continuousBreadcrumb(
	address string,
	breadcrumb *accountBreadcrumb,
	accountState state.UserAccountHandler,
) bool {

	if breadcrumb.hasUnknownNonce() {
		// this might occur when an account only acts as a relayer (never sender) in a specific tracked block.
		// in that case, we don't have any nonce info for the relayer.
		// as a result, its breadcrumb is treated as continuous.
		log.Debug("breadcrumbsValidator.continuousBreadcrumb breadcrumb has unknown nonce")
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
	if ok {
		return true
	}

	continuousWithSessionNonce := breadcrumb.verifyContinuityWithSessionNonce(accountNonce)
	if !continuousWithSessionNonce {
		validator.skippedSenders[address] = struct{}{}
		log.Debug("virtualSessionProvider.continuousBreadcrumb breadcrumb not continuous with session nonce",
			"address", address,
			"accountNonce", accountNonce,
			"breadcrumb nonce", breadcrumb.initialNonce)
		return false
	}
	validator.sendersInContinuityWithSessionNonce[address] = struct{}{}

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

func (validator *breadcrumbsValidator) shouldSkipSender(address string) bool {
	_, ok := validator.skippedSenders[address]
	return ok
}

func (validator *breadcrumbsValidator) validateBalance(
	address string,
	breadcrumb *accountBreadcrumb,
	initialBalance *big.Int,
) error {
	virtualBalance, ok := validator.virtualBalancesByAddress[address]
	if !ok {
		virtualBalance = newVirtualAccountBalance(initialBalance)
		validator.virtualBalancesByAddress[address] = virtualBalance
	}

	virtualBalance.accumulateConsumedBalance(breadcrumb)
	return virtualBalance.validateBalance()
}
