package bn

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

const (
	// MtUnknown defines ID of a message that has unknown Data inside
	MtUnknown MessageType = iota
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtCommitmentHash defines ID of a message that has a commitment hash inside
	MtCommitmentHash
	// MtBitmap defines ID of a message that has a bitmap inside
	MtBitmap
	// MtCommitment defines ID of a message that has a commitment inside
	MtCommitment
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
)

// syncThesholdPercent sepcifies the max allocated time to syncronize as a percentage of the total time of the round
const syncThresholdPercent = 50

// processingThresholdPercent specifies the max allocated time for processing the block as a percentage of the total time of the round
const processingThresholdPercent = 65

// maxThresholdPercent specifies the max allocated time percent for doing job as a percentage of the total time of one round
const maxThresholdPercent = 75

// srStartStartTime specifies the start time, from the total time of the round, of subround Start
const srStartStartTime = 0.0

// srEndStartTime specifies the end time, from the total time of the round, of subround Start
const srStartEndTime = 0.05

// srBlockStartTime specifies the start time, from the total time of the round, of subround Block
const srBlockStartTime = 0.05

// srBlockEndTime specifies the end time, from the total time of the round, of subround Block
const srBlockEndTime = 0.25

// srCommitmentHashStartTime specifies the start time, from the total time of the round, of subround CommitmentHash
const srCommitmentHashStartTime = 0.25

// srCommitmentHashEndTime specifies the end time, from the total time of the round, of subround CommitmentHash
const srCommitmentHashEndTime = 0.35

// srBitmapStartTime specifies the start time, from the total time of the round, of subround Bitmap
const srBitmapStartTime = 0.35

// srBitmapEndTime specifies the end time, from the total time of the round, of subround Bitmap
const srBitmapEndTime = 0.45

// srCommitmentStartTime specifies the start time, from the total time of the round, of subround Commitment
const srCommitmentStartTime = 0.45

// srCommitmentEndTime specifies the end time, from the total time of the round, of subround Commitment
const srCommitmentEndTime = 0.55

// srSignatureStartTime specifies the start time, from the total time of the round, of subround Signature
const srSignatureStartTime = 0.55

// srSignatureEndTime specifies the end time, from the total time of the round, of subround Signature
const srSignatureEndTime = 0.65

// srEndStartTime specifies the start time, from the total time of the round, of subround End
const srEndStartTime = 0.65

// srEndEndTime specifies the end time, from the total time of the round, of subround End
const srEndEndTime = 0.75
