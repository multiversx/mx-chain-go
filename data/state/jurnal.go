package state

import (
	"sync"
)

// Jurnal will keep track of all JurnalEntry objects
type Jurnal struct {
	accounts AccountsHandler
	entries  []JurnalEntry

	mutDirtyAddress sync.RWMutex
	dirtyAddresses  map[*Address]uint32
}

// NewJurnal creates a new Jurnal
func NewJurnal(accounts AccountsHandler) *Jurnal {
	j := Jurnal{
		accounts:        accounts,
		entries:         make([]JurnalEntry, 0),
		mutDirtyAddress: sync.RWMutex{},
		dirtyAddresses:  make(map[*Address]uint32),
	}

	return &j
}

// AddEntry adds a new object to entries list.
// Concurrent safe.
func (j *Jurnal) AddEntry(je JurnalEntry) {
	j.mutDirtyAddress.Lock()
	defer j.mutDirtyAddress.Unlock()

	j.entries = append(j.entries, je)
	val, ok := j.dirtyAddresses[je.DirtyAddress()]
	if !ok {
		val = 0
	}
	j.dirtyAddresses[je.DirtyAddress()] = val + 1
}

// RevertFromSnapshot apply Revert method over accounts object and removes it from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 1 index based. Calling this method with 0 will do nothing.
// Concurrent safe.
func (j *Jurnal) RevertFromSnapshot(snapshot uint32) error {
	if snapshot > uint32(len(j.entries)) || snapshot == 0 {
		//outside of bounds array, not quite error, just do NOP
		return nil
	}

	j.mutDirtyAddress.Lock()
	defer j.mutDirtyAddress.Unlock()

	for i := uint32(len(j.entries)); i >= snapshot; i-- {
		err := j.entries[i-1].Revert(j.accounts)

		if err != nil {
			return err
		}

		j.dirtyAddresses[j.entries[i-1].DirtyAddress()]--
		if j.dirtyAddresses[j.entries[i-1].DirtyAddress()] == 0 {
			delete(j.dirtyAddresses, j.entries[i-1].DirtyAddress())
		}
	}

	j.entries = j.entries[:snapshot-1]

	return nil
}

// Len will return the number of entries
// Concurrent safe.
func (j *Jurnal) Len() uint32 {
	j.mutDirtyAddress.RLock()
	defer j.mutDirtyAddress.RUnlock()

	return uint32(len(j.entries))
}
