package disabled

// Rater implements the Rater interface, it does nothing as it is disabled
type Rater struct {
}

// GetChance does nothing as it is disabled
func (r *Rater) GetChance(_ uint32) uint32 {
	return 0
}

// IsInterfaceNil returns true if underlying object is nil
func (r *Rater) IsInterfaceNil() bool {
	return r == nil
}
