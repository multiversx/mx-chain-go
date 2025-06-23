package txcache

type virtualSelectionSession struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

type virtualAccountRecord struct {
}

func newVirtualSelectionSession(session SelectionSession) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}
