package spos

//TODO consider moving these constants in config file

// MaxThresholdPercent specifies the max allocated time percent for doing Job as a percentage of the total time of one round
const MaxThresholdPercent = 95

// LeaderPeerHonestyIncreaseFactor specifies the factor with which the honesty of the leader should be increased
// if it proposed a block or sent the final info, in its correct allocated slot/time-frame/round
const LeaderPeerHonestyIncreaseFactor = 2

// ValidatorPeerHonestyIncreaseFactor specifies the factor with which the honesty of the validator should be increased
// if it sent the signature, in its correct allocated slot/time-frame/round
const ValidatorPeerHonestyIncreaseFactor = 1

// LeaderPeerHonestyDecreaseFactor specifies the factor with which the honesty of the leader should be decreased
// if it proposed a block or sent the final info, in an incorrect allocated slot/time-frame/round
const LeaderPeerHonestyDecreaseFactor = -4

// ValidatorPeerHonestyDecreaseFactor specifies the factor with which the honesty of the validator should be decreased
// if it sent the signature, in an incorrect allocated slot/time-frame/round
const ValidatorPeerHonestyDecreaseFactor = -2
