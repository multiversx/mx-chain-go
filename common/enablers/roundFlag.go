package enablers

import (
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
)

type roundFlag struct {
	*atomic.Flag
	round   uint64
	options []string
}
