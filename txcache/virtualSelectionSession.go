package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type virtualSelectionSession struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

type virtualAccountRecord struct {
	initialNonce   core.OptionalUint64
	initialBalance *big.Int
}

func newVirtualAccountRecord(initialNonce core.OptionalUint64, initialBalance *big.Int) *virtualAccountRecord {
	return &virtualAccountRecord{
		initialNonce:   initialNonce,
		initialBalance: initialBalance,
	}
}

func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	_ = virtualRecord.initialBalance.Add(virtualRecord.initialBalance, breadcrumb.consumedBalance)

	if !virtualRecord.initialNonce.HasValue {
		virtualRecord.initialNonce = breadcrumb.initialNonce
		return
	}

	if breadcrumb.initialNonce.HasValue {
		virtualRecord.initialNonce = core.OptionalUint64{
			Value:    max(breadcrumb.initialNonce.Value, virtualRecord.initialNonce.Value),
			HasValue: true,
		}
	}
}

func newVirtualSelectionSession(session SelectionSession) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}
