package state

import (
	"sync"
)

// Journal will keep track of all JournalEntry objects
type Journal struct {
	entries []JournalEntry

	mutDirtyAddress sync.RWMutex
}

// NewJournal creates a new Journal
func NewJournal() *Journal {
	return &Journal{
		entries:         make([]JournalEntry, 0),
		mutDirtyAddress: sync.RWMutex{},
	}
}

// AddEntry adds a new object to entries list.
// Concurrent safe.
func (j *Journal) AddEntry(je JournalEntry) {
	if je == nil {
		return
	}

	j.mutDirtyAddress.Lock()
	j.entries = append(j.entries, je)
	j.mutDirtyAddress.Unlock()
}

// RevertToSnapshot apply Revert method over accounts object and removes entries from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 0 index based. Calling this method with negative value will do nothing. Calling with 0 revert everything.
// Concurrent safe.
func (j *Journal) RevertToSnapshot(snapshot int, accountsAdapter AccountsAdapter) error {
	if snapshot > len(j.entries) || snapshot < 0 {
		//outside of bounds array, not quite error, just return
		return nil
	}

	j.mutDirtyAddress.Lock()
	defer j.mutDirtyAddress.Unlock()

	for i := len(j.entries) - 1; i >= snapshot; i-- {
		err := j.entries[i].Revert(accountsAdapter)

		if err != nil {
			return err
		}
	}

	j.entries = j.entries[:snapshot]

	return nil
}

// Len will return the number of entries
// Concurrent safe.
func (j *Journal) Len() int {
	j.mutDirtyAddress.RLock()
	length := len(j.entries)
	j.mutDirtyAddress.RUnlock()

	return length
}

// Clear clears the data from this journal.
func (j *Journal) Clear() {
	j.mutDirtyAddress.Lock()
	j.entries = make([]JournalEntry, 0)
	j.mutDirtyAddress.Unlock()
}

// Entries returns the entries saved in the journal
func (j *Journal) Entries() []JournalEntry {
	j.mutDirtyAddress.RLock()
	entries := j.entries
	j.mutDirtyAddress.RUnlock()

	return entries
}
