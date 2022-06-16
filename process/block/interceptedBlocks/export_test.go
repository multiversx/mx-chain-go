package interceptedBlocks

// IsMetaHeaderOutOfRange -
func (imh *InterceptedMetaHeader) IsMetaHeaderOutOfRange() bool {
	return imh.isMetaHeaderEpochOutOfRange()
}
