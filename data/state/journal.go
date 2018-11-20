package state

import (
	"sync"
)

// Journal will keep track of all JournalEntry objects
type Journal struct {
	accounts AccountsHandler
	entries  []JournalEntry

	mutDirtyAddress sync.RWMutex
	dirtyAddresses  map[*Address]int
}

// NewJournal creates a new Journal
func NewJournal(accounts AccountsHandler) *Journal {
	j := Journal{
		accounts:        accounts,
		entries:         make([]JournalEntry, 0),
		mutDirtyAddress: sync.RWMutex{},
		dirtyAddresses:  make(map[*Address]int),
	}

	return &j
}

// AddEntry adds a new object to entries list.
// Concurrent safe.
func (j *Journal) AddEntry(je JournalEntry) {
	j.mutDirtyAddress.Lock()
	defer j.mutDirtyAddress.Unlock()

	j.entries = append(j.entries, je)
	val := j.dirtyAddresses[je.DirtyAddress()]
	j.dirtyAddresses[je.DirtyAddress()] = val + 1
}

// RevertFromSnapshot apply Revert method over accounts object and removes it from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 0 index based. Calling this method with negative value will do nothing. Calling with 0 revert everything.
// Concurrent safe.
func (j *Journal) RevertFromSnapshot(snapshot int) error {
	if snapshot > len(j.entries) || snapshot < 0 {
		//outside of bounds array, not quite error, just do NOP
		return nil
	}

	j.mutDirtyAddress.Lock()
	defer j.mutDirtyAddress.Unlock()

	for i := len(j.entries) - 1; i >= snapshot; i-- {
		err := j.entries[i].Revert(j.accounts)

		if err != nil {
			return err
		}

		j.dirtyAddresses[j.entries[i].DirtyAddress()]--
		if j.dirtyAddresses[j.entries[i].DirtyAddress()] == 0 {
			delete(j.dirtyAddresses, j.entries[i].DirtyAddress())
		}
	}

	j.entries = j.entries[:snapshot]

	return nil
}

// Len will return the number of entries
// Concurrent safe.
func (j *Journal) Len() int {
	j.mutDirtyAddress.RLock()
	defer j.mutDirtyAddress.RUnlock()

	return len(j.entries)
}

// Clears the data from this journal.
func (j *Journal) Clear() {
	j.mutDirtyAddress.RLock()
	defer j.mutDirtyAddress.RUnlock()

	j.entries = make([]JournalEntry, 0)
	j.dirtyAddresses = make(map[*Address]int)
}

// Entries returns the entries saved in the journal
func (j *Journal) Entries() []JournalEntry {
	j.mutDirtyAddress.RLock()
	defer j.mutDirtyAddress.RUnlock()

	return j.entries
}
