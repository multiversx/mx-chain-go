package testscommon

// HardforkExclusionHandlerStub -
type HardforkExclusionHandlerStub struct {
	IsRoundExcludedCalled     func(round uint64) bool
	IsRollbackForbiddenCalled func(round uint64) bool
}

// IsRoundExcluded -
func (stub *HardforkExclusionHandlerStub) IsRoundExcluded(round uint64) bool {
	if stub.IsRoundExcludedCalled != nil {
		return stub.IsRoundExcludedCalled(round)
	}
	return false
}

// IsRollbackForbidden -
func (stub *HardforkExclusionHandlerStub) IsRollbackForbidden(round uint64) bool {
	if stub.IsRollbackForbiddenCalled != nil {
		return stub.IsRollbackForbiddenCalled(round)
	}
	return false
}

// IsInterfaceNil -
func (stub *HardforkExclusionHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
