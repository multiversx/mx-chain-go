package redundancy

// NewNilNodeRedundancy -
func NewNilNodeRedundancy() *nodeRedundancy {
	return nil
}

// GetRoundsOfInactivity -
func (nr *nodeRedundancy) GetRoundsOfInactivity() int {
	return nr.handler.RoundsOfInactivity()
}

// SetRoundsOfInactivity -
func (nr *nodeRedundancy) SetRoundsOfInactivity(roundsOfInactivity int) {
	nr.handler.ResetRoundsOfInactivity()

	for i := 0; i < roundsOfInactivity; i++ {
		nr.handler.IncrementRoundsOfInactivity()
	}
}

// GetLastRoundIndexCheck -
func (nr *nodeRedundancy) GetLastRoundIndexCheck() int64 {
	return nr.lastRoundIndexCheck
}

// SetLastRoundIndexCheck -
func (nr *nodeRedundancy) SetLastRoundIndexCheck(lastRoundIndexCheck int64) {
	nr.lastRoundIndexCheck = lastRoundIndexCheck
}
