package txpool

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type immunityCache interface {
	storage.Cacher
	ImmunizeTxsAgainstEviction(keys [][]byte)
}
