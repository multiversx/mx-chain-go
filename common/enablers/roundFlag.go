package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

type roundFlag struct {
	*atomic.Flag
	round   uint64
	options []string
}
