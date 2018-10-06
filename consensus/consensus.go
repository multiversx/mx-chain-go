package consensus

type Consensus struct {
	AgreementType
	ResponseType
	Threshold
}

func New() Consensus {
	c := Consensus{AgreementType: AT_NONE, ResponseType: RT_NONE}
	return c
}

// An AgreementType specifies the type of consensus (PBFT = 2/3 + 1)
type AgreementType int

const (
	AT_NONE AgreementType = iota
	AT_ONE
	AT_PBFT
	AT_ALL
)

// An ResponseType specifies a response of the node
type ResponseType int

const (
	// the validator did not answer (yet)
	RT_NONE ResponseType = iota
	// the validator agrees on consensus
	RT_AGREE
	// the validator disagree on consensus
	RT_DISAGREE
	// the validator is not available
	RT_NOT_AVAILABLE
)

type Threshold struct {
	Block         int
	ComitmentHash int
	Bitmap        int
	Comitment     int
	Signature     int
}
