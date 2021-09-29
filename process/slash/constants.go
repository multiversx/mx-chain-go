package slash

type SlashingType string

const (
	None             SlashingType = "none"
	MultipleProposal SlashingType = "multiple header proposal"
	MultipleSigning  SlashingType = "multiple header signing"
)

type SlashingLevel uint16

const (
	Level0 SlashingLevel = 0
	Level1 SlashingLevel = 1
	Level2 SlashingLevel = 2
)
