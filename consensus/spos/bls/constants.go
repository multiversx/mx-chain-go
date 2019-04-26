package bls

import "github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

const (
	// MtUnknown defines ID of a message that has unknown Data inside
	MtUnknown spos.MessageType = iota
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
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

// srSignatureStartTime specifies the start time, from the total time of the round, of subround Signature
const srSignatureStartTime = 0.25

// srSignatureEndTime specifies the end time, from the total time of the round, of subround Signature
const srSignatureEndTime = 0.65

// srEndStartTime specifies the start time, from the total time of the round, of subround End
const srEndStartTime = 0.65

// srEndEndTime specifies the end time, from the total time of the round, of subround End
const srEndEndTime = 0.75

func getStringValue(msgType spos.MessageType) string {
	switch msgType {
	case MtBlockBody:
		return "(BLOCK_BODY)"
	case MtBlockHeader:
		return "(BLOCK_HEADER)"
	case MtSignature:
		return "(SIGNATURE)"
	case MtUnknown:
		return "(UNKNOWN)"
	default:
		return "Undefined message type"
	}
}
