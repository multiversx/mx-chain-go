package slash

type SlashingType string

const (
	None             SlashingType = "none"
	DoubleProposal   SlashingType = "header double proposal"
	MultipleProposal SlashingType = "header multiple proposal"
)
