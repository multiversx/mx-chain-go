package txcache

import "math/big"

type virtualSelectionSession struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

type virtualAccountRecord struct {
	initialNonce   uint64
	initialBalance *big.Int
}

func newVirtualSelectionSession(session SelectionSession) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}
