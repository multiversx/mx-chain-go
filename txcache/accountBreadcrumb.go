package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/state"
)

type accountBreadcrumb struct {
	initialNonce    core.OptionalUint64
	lastNonce       core.OptionalUint64
	consumedBalance *big.Int
}

func newAccountBreadcrumb(
	initialNonce core.OptionalUint64,
	lastNonce core.OptionalUint64,
	consumedBalance *big.Int,
) *accountBreadcrumb {
	if consumedBalance == nil {
		consumedBalance = big.NewInt(0)
	}
	return &accountBreadcrumb{
		initialNonce:    initialNonce,
		lastNonce:       lastNonce,
		consumedBalance: consumedBalance,
	}
}

func (breadcrumb *accountBreadcrumb) accumulateConsumedBalance(transferredValue *big.Int) {
	if transferredValue != nil {
		_ = breadcrumb.consumedBalance.Add(breadcrumb.consumedBalance, transferredValue)
	}
}

func (breadcrumb *accountBreadcrumb) updateNonce(lastNonce core.OptionalUint64) {
	if lastNonce.HasValue {
		breadcrumb.lastNonce = lastNonce
	}
}

func (breadcrumb *accountBreadcrumb) createOrUpdateVirtualRecord(
	virtualAccountsByAddress map[string]*virtualAccountRecord,
	accountState state.UserAccountHandler,
	address string,
) {
	virtualRecord, ok := virtualAccountsByAddress[address]
	if !ok {
		initialBalance := accountState.GetBalance()
		virtualRecord = newVirtualAccountRecord(breadcrumb.initialNonce, initialBalance)
		virtualAccountsByAddress[address] = virtualRecord
	}

	virtualRecord.updateVirtualRecord(breadcrumb)
}

// TODO rename in case of adding a virtual selection session provider
func (breadcrumb *accountBreadcrumb) isContinuous(
	address string,
	accountNonce uint64,
	sendersInContinuityWithSessionNonce map[string]struct{},
	accountPreviousBreadcrumb map[string]*accountBreadcrumb,
) bool {
	if breadcrumb.hasUnkownNonce() {
		return true
	}

	_, ok := sendersInContinuityWithSessionNonce[address]
	continuousWithSessionNonce := breadcrumb.verifyContinuityWithSessionNonce(accountNonce)
	if !ok && !continuousWithSessionNonce {
		log.Debug("accountBreadcrumb.isContinuous breadcrumb not continuous with session nonce",
			"address", address,
			"accountNonce", accountNonce,
			"breadcrumb nonce", breadcrumb.initialNonce)
		return false
	}
	if !ok {
		sendersInContinuityWithSessionNonce[address] = struct{}{}
	}

	previousBreadcrumb := accountPreviousBreadcrumb[address]
	continuousBreadcrumbs := breadcrumb.verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumb)
	if !continuousBreadcrumbs {
		log.Debug("accountBreadcrumb.isContinuous breadcrumb not continuous with previous breadcrumb",
			"address", address,
			"accountNonce", accountNonce,
			"current breadcrumb nonce", breadcrumb.initialNonce,
			"previous breadcrumb nonce", previousBreadcrumb.initialNonce)
		return false
	}

	accountPreviousBreadcrumb[address] = breadcrumb
	return true
}

func (breadcrumb *accountBreadcrumb) verifyContinuityBetweenAccountBreadcrumbs(
	previousBreadcrumbAsSender *accountBreadcrumb,
) bool {
	return previousBreadcrumbAsSender == nil || previousBreadcrumbAsSender.lastNonce.Value+1 == breadcrumb.initialNonce.Value
}

func (breadcrumb *accountBreadcrumb) verifyContinuityWithSessionNonce(sessionNonce uint64) bool {
	return breadcrumb.initialNonce.Value == sessionNonce
}

func (breadcrumb *accountBreadcrumb) hasUnkownNonce() bool {
	return !breadcrumb.initialNonce.HasValue && !breadcrumb.lastNonce.HasValue
}
