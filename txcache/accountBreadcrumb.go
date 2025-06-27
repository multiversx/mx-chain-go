package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type accountBreadcrumb struct {
	initialNonce    core.OptionalUint64
	lastNonce       core.OptionalUint64
	consumedBalance *big.Int
}
