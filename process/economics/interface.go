package economics

type RatingsInfo interface {
	StartRating() uint32
	MaxRating() uint32
	MinRating() uint32
	ProposerIncreaseRatingStep() uint32
	ProposerDecreaseRatingStep() uint32
	ValidatorIncreaseRatingStep() uint32
	ValidatorDecreaseRatingStep() uint32
	SelectionChances() []SelectionChance
}
