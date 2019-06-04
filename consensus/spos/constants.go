package spos

// maxThresholdPercent specifies the max allocated time percent for doing Job as a percentage of the total time of one round
const maxThresholdPercent = 75

// MaxRoundsGap defines the maximum expected gap in terms of rounds, between metachain and shardchain, after which
// a block committed and broadcast from shardchain would be visible as notarized in metachain
const MaxRoundsGap = 3
