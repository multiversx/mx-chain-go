package txcache

import "math/big"

type virtualAccountBalance struct {
	initialBalance  *big.Int
	consumedBalance *big.Int
}

func newVirtualAccountBalance(initialBalance *big.Int) *virtualAccountBalance {
	return &virtualAccountBalance{
		initialBalance:  initialBalance,
		consumedBalance: big.NewInt(0),
	}
}

func (virtualBalance *virtualAccountBalance) accumulateConsumedBalance(breadcrumb *accountBreadcrumb) {
	_ = virtualBalance.consumedBalance.Add(virtualBalance.consumedBalance, breadcrumb.consumedBalance)
}

func (virtualBalance *virtualAccountBalance) validateBalance() error {
	if virtualBalance.consumedBalance.Cmp(virtualBalance.initialBalance) > 0 {
		return errExceedBalance
	}

	return nil
}

func (virtualBalance *virtualAccountBalance) getInitialBalance() *big.Int {
	return virtualBalance.initialBalance
}

func (virtualBalance *virtualAccountBalance) getConsumedBalance() *big.Int {
	return virtualBalance.consumedBalance
}
