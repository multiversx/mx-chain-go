package chronology

// A RoundState specifies in which state is the current round
type RoundState int

const (
	RS_START_ROUND RoundState = iota
	RS_PROPOSE_BLOCK
	RS_END_ROUND
)
