package slash

type SlashingType string

const (
	None             SlashingType = "none"
	MultipleProposal SlashingType = "header multiple proposal"
)

type SlashingLevel uint16

const (
	Level0 SlashingLevel = 0
	Level1 SlashingLevel = 1
	Level2 SlashingLevel = 2
)
