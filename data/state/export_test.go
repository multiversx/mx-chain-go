package state

func (j *Jurnal) Entries() []JurnalEntry {
	return j.entries
}

func (j *Jurnal) DirtyAddresses() map[*Address]uint32 {
	return j.dirtyAddresses
}
