package enableRounds

import "github.com/ElrondNetwork/elrond-go-core/core/atomic"

type roundFlag struct {
	round   uint64
	flag    *atomic.Flag
	options []string
}
