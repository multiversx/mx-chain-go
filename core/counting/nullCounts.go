package counting

var _ Counts = (*NullCounts)(nil)

// NullCounts implements null object pattern for counts
type NullCounts struct {
}

// GetTotal returns -1
func (counts *NullCounts) GetTotal() int64 {
	return -1
}

// GetTotalSize returns -1
func (counts *NullCounts) GetTotalSize() int64 {
	return -1
}

// String returns a placeholder
func (counts *NullCounts) String() string {
	return "counts not applicable"
}

// IsInterfaceNil returns true if there is no value under the interface
func (counts *NullCounts) IsInterfaceNil() bool {
	return counts == nil
}
