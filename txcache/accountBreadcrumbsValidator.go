package txcache

import (
	"math/big"
)

// breadcrumbsValidator checks that breadcrumbs are continuous:
// With the session nonce.
// With the previous breadcrumbs.
type breadcrumbsValidator struct {
	skippedSenders                      map[string]struct{}
	sendersInContinuityWithSessionNonce map[string]struct{}
	accountPreviousBreadcrumb           map[string]*accountBreadcrumb
	virtualBalancesByAddress            map[string]*virtualAccountBalance
}

// newBreadcrumbValidator is used when the validation for a proposed block is called (when receiving the OnProposedBlock notification).
// At the end of the method it becomes useless.
func newBreadcrumbValidator() *breadcrumbsValidator {
	return &breadcrumbsValidator{
		skippedSenders:                      make(map[string]struct{}),
		sendersInContinuityWithSessionNonce: make(map[string]struct{}),
		accountPreviousBreadcrumb:           make(map[string]*accountBreadcrumb),
		virtualBalancesByAddress:            make(map[string]*virtualAccountBalance),
	}
}

// validateNonceContinuityOfBreadcrumb is used when a block is proposed.
func (validator *breadcrumbsValidator) validateNonceContinuityOfBreadcrumb(
	address string,
	accountSessionNonce uint64,
	breadcrumb *accountBreadcrumb,
) bool {
	if breadcrumb.hasUnknownNonce() {
		// this might occur when an account only acts as a relayer (never sender) in a specific tracked block.
		// in that case, we don't have any nonce info for the relayer.
		// as a result, its breadcrumb is treated as continuous.
		log.Trace("breadcrumbsValidator.validateNonceContinuityOfBreadcrumb breadcrumb has unknown nonce")
		return true
	}

	if !validator.validateContinuityWithSessionNonce(address, accountSessionNonce, breadcrumb) {
		return false
	}

	if !validator.validateContinuityWithPreviousBreadcrumb(address, breadcrumb) {
		return false
	}

	return true
}

// validateContinuityWithSessionNonce checks that the given breadcrumb is continuous with the session nonce
func (validator *breadcrumbsValidator) validateContinuityWithSessionNonce(
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
		// mark this sender as not continuous
		validator.skippedSenders[address] = struct{}{}
		log.Debug("virtualSessionComputer.validateNonceContinuityOfBreadcrumb breadcrumb not continuous with session nonce",
			"address", address,
			"accountNonce", accountNonce,
			"breadcrumb nonce", breadcrumb.firstNonce)
		return false
	}

	// mark this sender as continuous with the session nonce
	validator.sendersInContinuityWithSessionNonce[address] = struct{}{}
	return true
}

// validateContinuityWithPreviousBreadcrumb checks that the given breadcrumb of the address is continuous with the previous one, saved in the internal state of the validator.
func (validator *breadcrumbsValidator) validateContinuityWithPreviousBreadcrumb(
	address string,
	breadcrumb *accountBreadcrumb) bool {

	previousBreadcrumb := validator.accountPreviousBreadcrumb[address]
	continuousBreadcrumbs := breadcrumb.verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumb)
	if !continuousBreadcrumbs {
		// mark this sender as not continuous
		validator.skippedSenders[address] = struct{}{}
		log.Debug("virtualSessionComputer.validateNonceContinuityOfBreadcrumb breadcrumb not continuous with previous breadcrumb",
			"address", address,
			"current breadcrumb nonce", breadcrumb.firstNonce,
			"previous breadcrumb nonce", previousBreadcrumb.lastNonce)
		return false
	}

	// update the previous breadcrumb of this address
	validator.accountPreviousBreadcrumb[address] = breadcrumb

	return true
}

func (validator *breadcrumbsValidator) shouldSkipSender(address string) bool {
	_, ok := validator.skippedSenders[address]
	return ok
}

// validateBalance is used for the OnProposedBlock flow, when validating the compiled breadcrumbs.
func (validator *breadcrumbsValidator) validateBalance(
	address string,
	initialBalance *big.Int,
	breadcrumb *accountBreadcrumb,
) error {
	virtualBalance, ok := validator.virtualBalancesByAddress[address]
	if !ok {
		balance, err := newVirtualAccountBalance(initialBalance)
		if err != nil {
			return err
		}

		virtualBalance = balance
		validator.virtualBalancesByAddress[address] = balance
	}

	virtualBalance.accumulateConsumedBalance(breadcrumb.consumedBalance)
	return virtualBalance.validateBalance()
}
