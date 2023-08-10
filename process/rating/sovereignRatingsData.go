package rating

import "github.com/multiversx/mx-chain-go/process"

type sovereignRatingsData struct {
	*RatingsData
}

// NewSovereignRatingsData creates a sovereign ratings data
func NewSovereignRatingsData(args RatingsDataArg) (*sovereignRatingsData, error) {
	ratingsConfig := args.Config
	err := verifyRatingsConfig(ratingsConfig)
	if err != nil {
		return nil, err
	}

	chances := make([]process.SelectionChance, 0)
	for _, chance := range ratingsConfig.General.SelectionChances {
		chances = append(chances, &SelectionChance{
			MaxThreshold:  chance.MaxThreshold,
			ChancePercent: chance.ChancePercent,
		})
	}

	arg := computeRatingStepArg{
		shardSize:                       args.ShardMinNodes,
		consensusSize:                   args.ShardConsensusSize,
		roundTimeMilis:                  args.RoundDurationMiliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: ratingsConfig.ShardChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          ratingsConfig.ShardChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         ratingsConfig.ShardChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     ratingsConfig.ShardChain.ProposerValidatorImportance,
	}
	shardRatingStep, err := computeRatingStep(arg)
	if err != nil {
		return nil, err
	}

	return &sovereignRatingsData{
		RatingsData: &RatingsData{
			startRating:           ratingsConfig.General.StartRating,
			maxRating:             ratingsConfig.General.MaxRating,
			minRating:             ratingsConfig.General.MinRating,
			signedBlocksThreshold: ratingsConfig.General.SignedBlocksThreshold,
			metaRatingsStepData:   shardRatingStep,
			shardRatingsStepData:  shardRatingStep,
			selectionChances:      chances,
		},
	}, nil
}
