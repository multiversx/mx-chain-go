package consensus

type Consensus struct {
	Agreement AgreementType
	Answer    AnswerType
}

// An AgreementType specifies the type of consensus (PBFT = 2/3 + 1)
type AgreementType int

const (
	AT_NONE AgreementType = iota
	AT_ONE
	AT_PBFT
	AT_ALL
)

// An AnswerType specifies an answer of the node
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
