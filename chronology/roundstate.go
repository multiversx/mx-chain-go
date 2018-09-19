package chronology

// A RoundState specifies in which state is the current round
type RoundState int

const (
	RS_BEFORE_ROUND RoundState = iota
	RS_START_ROUND
	RS_PROPOSE_BLOCK
	RS_SEND_COMITMENT_HASH
	RS_SEND_BITMAP
	RS_SEND_COMITMENT
	RS_SEND_AGGREGATE_COMITMENT
	RS_END_ROUND
	RS_AFTER_ROUND
	RS_UNKNOWN
)
