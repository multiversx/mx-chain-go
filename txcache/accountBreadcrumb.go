package txcache

import "math/big"

type accountBreadcrumb struct {
	initialNonce    uint64
	lastNonce       uint64
	consumedBalance *big.Int
}
