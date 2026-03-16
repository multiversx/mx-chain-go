package v2

import (
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("consensus/spos/bls/v2")

// waitingAllSigsMaxTimeThreshold specifies the max allocated time for waiting all signatures from the total time of the subround signature
const waitingAllSigsMaxTimeThreshold = 0.5

// processingThresholdPercent specifies the max allocated time for processing the block as a percentage of the total time of the round
const processingThresholdPercent = 85

// srStartStartTime specifies the start time, from the total time of the round, of Subround Start
const srStartStartTime = 0.0

// srEndStartTime specifies the end time, from the total time of the round, of Subround Start
const srStartEndTime = 0.05

// srBlockStartTime specifies the start time, from the total time of the round, of Subround Block
const srBlockStartTime = 0.05

// srBlockEndTime specifies the end time, from the total time of the round, of Subround Block
const srBlockEndTime = 0.25

// srSignatureStartTime specifies the start time, from the total time of the round, of Subround Signature
const srSignatureStartTime = 0.25

// srSignatureEndTime specifies the end time, from the total time of the round, of Subround Signature
const srSignatureEndTime = 0.85

// srEndStartTime specifies the start time, from the total time of the round, of Subround End
const srEndStartTime = 0.85

// srEndEndTime specifies the end time, from the total time of the round, of Subround End
const srEndEndTime = 0.95
