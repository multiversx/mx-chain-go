package bootstrap

// bootstrapEpochProvider keeps main-network requests enabled during bootstrap.
// Historical full-archive bootstrap also queries archive peers: the selected
// epoch can reference data older than regular peers retain. This uses the
// request sender's existing dual-network policy and bounded peer counts.
type bootstrapEpochProvider struct {
	queryFullArchive bool
}

func (b *bootstrapEpochProvider) EpochIsActiveInNetwork(_ uint32) bool {
	return !b.queryFullArchive
}

func (b *bootstrapEpochProvider) EpochIsAvailableOnMainPeers(_ uint32) bool {
	return true
}

func (b *bootstrapEpochProvider) EpochConfirmed(_ uint32, _ uint64) {
}

func (b *bootstrapEpochProvider) IsInterfaceNil() bool {
	return b == nil
}
