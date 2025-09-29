package txcache

import (
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

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
	gab.setForFeePayer()

	return gab
}

// setForFeePayer sets a global account breadcrumb for a fee payer
func (gab *globalAccountBreadcrumb) setForFeePayer() {
	gab.firstNonce = core.OptionalUint64{
		Value:    math.MaxUint64,
		HasValue: false,
	}
	gab.lastNonce = core.OptionalUint64{
		Value:    0,
		HasValue: false,
	}
}

// updateOnAddedAccountBreadcrumb updates a global breadcrumb when a tracked block is added
func (gab *globalAccountBreadcrumb) updateOnAddedAccountBreadcrumb(receivedBreadcrumb *accountBreadcrumb) {
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

	gab.extendRightRange(receivedBreadcrumb)
}

// extendRightRange extends the nonce range by maximizing the last nonce
func (gab *globalAccountBreadcrumb) extendRightRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.lastNonce.Value = max(gab.lastNonce.Value, receivedBreadcrumb.lastNonce.Value)
	gab.lastNonce.HasValue = true
}

// updateOnRemoveAccountBreadcrumbOnExecutedBlock updates the global account breadcrumb when a block is removed on the OnExecutedBlock notification
func (gab *globalAccountBreadcrumb) updateOnRemoveAccountBreadcrumbOnExecutedBlock(receivedBreadcrumb *accountBreadcrumb) (bool, error) {
	err := gab.reduceConsumedBalance(receivedBreadcrumb)
	if err != nil {
		return false, err
	}

	hasSameLastNonce := gab.lastNonce == receivedBreadcrumb.lastNonce

	// if our global breadcrumb has same last nonce with the received one it means we can mark it as a fee payer
	if !gab.isFeePayer() && hasSameLastNonce {
		gab.setForFeePayer()
	}

	if !gab.isFeePayer() {
		if receivedBreadcrumb.hasUnknownNonce() {
			// we do not update with nonce info from breadcrumb of relayer
			return false, nil
		}

		gab.reduceLeftRange(receivedBreadcrumb)
	}

	return gab.canBeDeleted(), nil
}

// extendRightRange reduces the nonce range by minimizing the first nonce
func (gab *globalAccountBreadcrumb) reduceLeftRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.firstNonce.Value = max(gab.firstNonce.Value, receivedBreadcrumb.lastNonce.Value+1)
	gab.firstNonce.HasValue = true
}

// updateOnRemoveAccountBreadcrumbOnExecutedBlock updates the global account breadcrumb when a block is removed on the OnProposedBlock notification
func (gab *globalAccountBreadcrumb) updateOnRemoveAccountBreadcrumbOnProposedBlock(receivedBreadcrumb *accountBreadcrumb) (bool, error) {
	err := gab.reduceConsumedBalance(receivedBreadcrumb)
	if err != nil {
		return false, err
	}

	hasSameFirstNonce := gab.firstNonce == receivedBreadcrumb.firstNonce

	// if our global breadcrumb has same last nonce with the received one it means we can mark it as a fee payer
	if !gab.isFeePayer() && hasSameFirstNonce {
		gab.setForFeePayer()
	}

	if !gab.isFeePayer() {
		if receivedBreadcrumb.hasUnknownNonce() {
			// we do not update with nonce info from breadcrumb of relayer
			return false, nil
		}

		gab.reduceRightRange(receivedBreadcrumb)
	}

	return gab.canBeDeleted(), nil
}

// reduceRightRange reduces the nonce range by minimizing the last nonce
func (gab *globalAccountBreadcrumb) reduceRightRange(receivedBreadcrumb *accountBreadcrumb) {
	gab.lastNonce.Value = min(gab.lastNonce.Value, receivedBreadcrumb.firstNonce.Value-1)
	gab.lastNonce.HasValue = true
}

// reduceConsumedBalance reduces the consumed balance of the global account breadcrumb by subtracting the received balance
func (gab *globalAccountBreadcrumb) reduceConsumedBalance(receivedBreadcrumb *accountBreadcrumb) error {
	_ = gab.consumedBalance.Sub(gab.consumedBalance, receivedBreadcrumb.consumedBalance)
	if gab.consumedBalance.Cmp(big.NewInt(0)) == -1 {
		return errNegativeBalance
	}

	return nil
}

// extendConsumedBalance accumulates the consumed balance of the global account breadcrumb by adding the received balance
func (gab *globalAccountBreadcrumb) extendConsumedBalance(receivedBreadcrumb *accountBreadcrumb) {
	// the consumed balance has to be updated despite the type of the breadcrumb (relayer or sender)
	_ = gab.consumedBalance.Add(gab.consumedBalance, receivedBreadcrumb.consumedBalance)
}

func (gab *globalAccountBreadcrumb) isFeePayer() bool {
	return !gab.firstNonce.HasValue || !gab.lastNonce.HasValue
}

func (gab *globalAccountBreadcrumb) canBeDeleted() bool {
	hasConsumedBalance := gab.consumedBalance.Cmp(big.NewInt(0)) == 1
	// it might be possible to delete that breadcrumb from the global map if it is a relayer breadcrumb but does not have a consumed balance
	if gab.isFeePayer() && !hasConsumedBalance {
		return true
	}

	return false
}
