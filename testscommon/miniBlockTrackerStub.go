package testscommon

// MiniBlockTrackerStub is a stub for process.MiniBlockTracker
type MiniBlockTrackerStub struct {
	ReleaseImmunityForCommittedMetaBlocksCalled  func(threshold uint64)
	ReleaseImmunityForCommittedShardBlocksCalled func(senderShard uint32, threshold uint64)
}

// ReleaseImmunityForCommittedMetaBlocks -
func (s *MiniBlockTrackerStub) ReleaseImmunityForCommittedMetaBlocks(threshold uint64) {
	if s.ReleaseImmunityForCommittedMetaBlocksCalled != nil {
		s.ReleaseImmunityForCommittedMetaBlocksCalled(threshold)
	}
}

// ReleaseImmunityForCommittedShardBlocks -
func (s *MiniBlockTrackerStub) ReleaseImmunityForCommittedShardBlocks(senderShard uint32, threshold uint64) {
	if s.ReleaseImmunityForCommittedShardBlocksCalled != nil {
		s.ReleaseImmunityForCommittedShardBlocksCalled(senderShard, threshold)
	}
}

// IsInterfaceNil returns true if the receiver is nil
func (s *MiniBlockTrackerStub) IsInterfaceNil() bool {
	return s == nil
}
