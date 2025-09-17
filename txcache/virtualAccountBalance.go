package txcache

import "math/big"

// virtualAccountBalance contains:
// the initialBalance from the non-virtual session,
// the consumedBalance accumulated from breadcrumbs.
type virtualAccountBalance struct {
	initialBalance  *big.Int
	consumedBalance *big.Int
}

// virtualAccountBalance is used in two scenarios:
// when validating a proposed block (on the OnProposedBlock notification), where we create a virtualBalance for each account;
// inside a virtual record
func newVirtualAccountBalance(initialBalance *big.Int) (*virtualAccountBalance, error) {
	if initialBalance == nil {
		return nil, errNilBalance
	}
	return &virtualAccountBalance{
		initialBalance:  initialBalance,
		consumedBalance: big.NewInt(0),
	}, nil
}

// accumulateConsumedBalance is used in two places:
// accumulating for the validation of a proposed block
// accumulating for a virtual record
func (virtualBalance *virtualAccountBalance) accumulateConsumedBalance(consumedBalance *big.Int) {
	_ = virtualBalance.consumedBalance.Add(virtualBalance.consumedBalance, consumedBalance)
}

// validateBalance is used in ONLY one place: the validation of a proposed block
// this method is NOT used for the virtual records (in deriveVirtualSelectionSession)
func (virtualBalance *virtualAccountBalance) validateBalance() error {
	if virtualBalance.consumedBalance.Cmp(virtualBalance.initialBalance) > 0 {
		return errExceededBalance
	}

	return nil
}

func (virtualBalance *virtualAccountBalance) getInitialBalance() *big.Int {
	return virtualBalance.initialBalance
}

func (virtualBalance *virtualAccountBalance) getConsumedBalance() *big.Int {
	return virtualBalance.consumedBalance
}
