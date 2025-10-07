package txcache

import (
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

// globalAccountBreadcrumb will be used for the elements stored in the globalAccountBreadcrumbs from the SelectionTracker.
// A globalAccountBreadcrumb should be affected each time a tracked block is added or removed.
type globalAccountBreadcrumb struct {
	firstNonce      core.OptionalUint64
	lastNonce       core.OptionalUint64
	consumedBalance *big.Int
}

// newGlobalAccountBreadcrumb creates a new global account breadcrumb
func newGlobalAccountBreadcrumb() *globalAccountBreadcrumb {
	gab := &globalAccountBreadcrumb{
		consumedBalance: big.NewInt(0),
	}
	gab.resetNonces()

	return gab
}

// resetNonces sets a global account breadcrumb for a fee payer
func (gab *globalAccountBreadcrumb) resetNonces() {
	gab.firstNonce = core.OptionalUint64{
		Value:    math.MaxUint64,
		HasValue: false,
	}
	gab.lastNonce = core.OptionalUint64{
		Value:    0,
		HasValue: false,
	}
}

// updateOnAddedBreadcrumb updates a global breadcrumb when a tracked block is added,
// but should be used only after the validation of the proposed block passed.
func (gab *globalAccountBreadcrumb) updateOnAddedBreadcrumb(receivedBreadcrumb *accountBreadcrumb) {
	gab.extendConsumedBalance(receivedBreadcrumb)

	// when a tracked block is added, we should only extend the nonce range of the global account breadcrumb
	if receivedBreadcrumb.hasUnknownNonce() {
		// we do not update with nonce info from breadcrumb of relayer
		return
	}

	// check if it is the first time as sender
	if !gab.firstNonce.HasValue {
		gab.firstNonce = receivedBreadcrumb.firstNonce
	}

	gab.extendRightNonceRange(receivedBreadcrumb)
}

// extendRightNonceRange extends the nonce range by maximizing the last nonce
func (gab *globalAccountBreadcrumb) extendRightNonceRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.lastNonce.Value = max(gab.lastNonce.Value, receivedBreadcrumb.lastNonce.Value)
	gab.lastNonce.HasValue = true
}

// updateOnRemovedBreadcrumbWithSameNonceOrBelow updates the global account breadcrumb when a block is removed on the OnExecutedBlock notification
func (gab *globalAccountBreadcrumb) updateOnRemovedBreadcrumbWithSameNonceOrBelow(receivedBreadcrumb *accountBreadcrumb) (bool, error) {
	err := gab.reduceConsumedBalance(receivedBreadcrumb)
	if err != nil {
		return false, err
	}

	hasSameLastNonce := gab.lastNonce == receivedBreadcrumb.lastNonce

	// if our global breadcrumb has same last nonce with the received one it means we can mark it as a fee payer
	if gab.isUser() && hasSameLastNonce {
		gab.resetNonces()
	}

	if gab.isUser() {
		if receivedBreadcrumb.hasUnknownNonce() {
			// should not update with nonce info from breadcrumb of relayer
			return false, nil
		}

		gab.reduceLeftNonceRange(receivedBreadcrumb)
	}

	return gab.canBeDeleted(), nil
}

// reduceLeftNonceRange reduces the left range by maximizing the first nonce
func (gab *globalAccountBreadcrumb) reduceLeftNonceRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.firstNonce.Value = max(gab.firstNonce.Value, receivedBreadcrumb.lastNonce.Value+1)
	gab.firstNonce.HasValue = true
}

// updateOnRemoveBreadcrumbWithSameNonceOrAbove updates the global account breadcrumb when a block is removed on the OnProposedBlock notification,
// but should be used only after the validation of the proposed block passed.
func (gab *globalAccountBreadcrumb) updateOnRemoveBreadcrumbWithSameNonceOrAbove(receivedBreadcrumb *accountBreadcrumb) (bool, error) {
	err := gab.reduceConsumedBalance(receivedBreadcrumb)
	if err != nil {
		return false, err
	}

	hasSameFirstNonce := gab.firstNonce == receivedBreadcrumb.firstNonce

	// if our global breadcrumb has same first nonce with the received one it means we can mark it as a fee payer
	if gab.isUser() && hasSameFirstNonce {
		gab.resetNonces()
	}

	if gab.isUser() {
		if receivedBreadcrumb.hasUnknownNonce() {
			// should not update with nonce info from breadcrumb of relayer
			return false, nil
		}

		gab.reduceRightNonceRange(receivedBreadcrumb)
	}

	return gab.canBeDeleted(), nil
}

// reduceRightNonceRange reduces the nonce range by minimizing the last nonce
func (gab *globalAccountBreadcrumb) reduceRightNonceRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.lastNonce.Value = min(gab.lastNonce.Value, receivedBreadcrumb.firstNonce.Value-1)
	gab.lastNonce.HasValue = true
}

// reduceConsumedBalance reduces the consumed balance of the global account breadcrumb by subtracting the received balance
func (gab *globalAccountBreadcrumb) reduceConsumedBalance(receivedBreadcrumb *accountBreadcrumb) error {
	_ = gab.consumedBalance.Sub(gab.consumedBalance, receivedBreadcrumb.consumedBalance)
	if gab.consumedBalance.Sign() == -1 {
		return errNegativeBalanceForBreadcrumb
	}

	return nil
}

// extendConsumedBalance accumulates the consumed balance of the global account breadcrumb by adding the received balance
func (gab *globalAccountBreadcrumb) extendConsumedBalance(receivedBreadcrumb *accountBreadcrumb) {
	// the consumed balance has to be updated despite the type of the breadcrumb (relayer or sender)
	_ = gab.consumedBalance.Add(gab.consumedBalance, receivedBreadcrumb.consumedBalance)
}

func (gab *globalAccountBreadcrumb) isUser() bool {
	return gab.firstNonce.HasValue || gab.lastNonce.HasValue
}

func (gab *globalAccountBreadcrumb) canBeDeleted() bool {
	hasConsumedBalance := gab.consumedBalance.Sign() == 1
	// it might be possible to delete the address of the breadcrumb from the global map,
	// but only if it is a relayer breadcrumb and its consumed balance is equal to 0.
	if !gab.isUser() && !hasConsumedBalance {
		return true
	}

	return false
}
