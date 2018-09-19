package consensus

// An Answer specifies an answer of the node
type AnswerType int

const (
	// the validator agrees on consensus
	AT_AGREE AnswerType = iota
	// the validator disagree on consensus
	AT_DISAGREE
	// the validator did not answer (yet)
	AT_NOT_ANSWERED
	// the validator is not available
	AT_NOT_AVAILABLE
)
