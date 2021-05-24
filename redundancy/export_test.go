package redundancy

// GetMaxRoundsOfInactivityAccepted -
func GetMaxRoundsOfInactivityAccepted() uint64 {
	return maxRoundsOfInactivityAccepted
}

// GetRoundsOfInactivity -
func (nr *nodeRedundancy) GetRoundsOfInactivity() uint64 {
	return nr.roundsOfInactivity
}

// SetRoundsOfInactivity -
func (nr *nodeRedundancy) SetRoundsOfInactivity(roundsOfInactivity uint64) {
	nr.roundsOfInactivity = roundsOfInactivity
}

// GetLastRoundIndexCheck -
func (nr *nodeRedundancy) GetLastRoundIndexCheck() int64 {
	return nr.lastRoundIndexCheck
}

// SetLastRoundIndexCheck -
func (nr *nodeRedundancy) SetLastRoundIndexCheck(lastRoundIndexCheck int64) {
	nr.lastRoundIndexCheck = lastRoundIndexCheck
}
