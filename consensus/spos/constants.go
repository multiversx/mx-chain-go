package spos

// MaxThresholdPercent specifies the max allocated time percent for doing Job as a percentage of the total time of one round
const MaxThresholdPercent = 95

// ProposerRatingIncreaseFactor specifies the factor with which the rating of a proposer should be increased if it
// proposed a block or sent final info, in its correct allocated slot/time-frame/round
const ProposerRatingIncreaseFactor = 2.0

// ValidatorRatingIncreaseFactor specifies the factor with which the rating of a validator should be increased if it
// sent its signature, in its correct allocated slot/time-frame/round
const ValidatorRatingIncreaseFactor = 1.0

// ProposerRatingDecreaseFactor specifies the factor with which the rating of a proposer should be decreased if it
// proposed a block or sent final info, in an incorrect allocated slot/time-frame/round
const ProposerRatingDecreaseFactor = -4.0

// ValidatorRatingDecreaseFactor specifies the factor with which the rating of a validator should be decreased if it
// sent its signature, in an incorrect allocated slot/time-frame/round
const ValidatorRatingDecreaseFactor = -2.0
