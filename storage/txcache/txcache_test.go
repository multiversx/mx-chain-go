package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

func TestPut(t *testing.T) {
	cache := NewTxCache(1000, 16)

	tx := &transaction.Transaction{}
	cache.AddTx([]byte("test"), tx)
}
