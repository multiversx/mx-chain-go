package counting

var _ Counts = (*NullCounts)(nil)

// NullCounts implements null object pattern for counts
type NullCounts struct {
}

// GetTotal gets total count
func (counts *NullCounts) GetTotal() int64 {
	return -1
}

func (counts *NullCounts) String() string {
	return "counts not applicable"
}
