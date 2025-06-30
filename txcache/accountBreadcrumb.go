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

func (breadcrumb *accountBreadcrumb) isContinuous(
	address string,
	accountNonce uint64,
	sendersInContinuityWithSessionNonce map[string]struct{},
	accountPreviousBreadcrumb map[string]*accountBreadcrumb,
) bool {
	if breadcrumb.isRelayer() {
		return true
	}

	_, ok := sendersInContinuityWithSessionNonce[address]
	if !ok && !breadcrumb.verifyContinuityWithSessionNonce(accountNonce) {
		return false
	} else if !ok {
		sendersInContinuityWithSessionNonce[address] = struct{}{}
	}

	previousBreadcrumb, ok := accountPreviousBreadcrumb[address]
	if ok &&
		!breadcrumb.verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumb) {
		return false
	}

	accountPreviousBreadcrumb[address] = breadcrumb
	return true
}

func (breadcrumb *accountBreadcrumb) verifyContinuityBetweenAccountBreadcrumbs(
	previousBreadcrumbAsSender *accountBreadcrumb,
) bool {
	return previousBreadcrumbAsSender.lastNonce.Value+1 == breadcrumb.initialNonce.Value
}

func (breadcrumb *accountBreadcrumb) verifyContinuityWithSessionNonce(sessionNonce uint64) bool {
	return breadcrumb.initialNonce.Value == sessionNonce
}

func (breadcrumb *accountBreadcrumb) isRelayer() bool {
	if !breadcrumb.initialNonce.HasValue && !breadcrumb.lastNonce.HasValue {
		return true
	}

	return false
}
