package state

import (
	"math/big"
)

// MetaJournalEntryTxCount is used to revert a tx count change
type MetaJournalEntryTxCount struct {
	jurnalizedAccount JournalizedAccountWrapper
	oldTxCount        *big.Int
}
