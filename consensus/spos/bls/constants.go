package bls

import (
	"github.com/multiversx/mx-chain-go/consensus"
)

const (
	// SrStartRound defines ID of Subround "Start round"
	SrStartRound = iota
	// SrBlock defines ID of Subround "block"
	SrBlock
	// SrSignature defines ID of Subround "signature"
	SrSignature
	// SrEndRound defines ID of Subround "End round"
	SrEndRound
)

const (
	// MtUnknown defines ID of a message that has unknown data inside
	MtUnknown consensus.MessageType = iota
	// MtBlockBodyAndHeader defines ID of a message that has a block body and a block header inside
	MtBlockBodyAndHeader
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
	// MtBlockHeaderFinalInfo defines ID of a message that has a block header final info inside
	// (aggregate signature, bitmap and seal leader signature for the proposed and accepted header)
	MtBlockHeaderFinalInfo
	// MtInvalidSigners defines ID of a message that has a invalid signers p2p messages inside
	MtInvalidSigners
)

const (
	// BlockBodyAndHeaderStringValue represents the string to be used to identify a block body and a block header
	BlockBodyAndHeaderStringValue = "(BLOCK_BODY_AND_HEADER)"

	// BlockBodyStringValue represents the string to be used to identify a block body
	BlockBodyStringValue = "(BLOCK_BODY)"

	// BlockHeaderStringValue represents the string to be used to identify a block header
	BlockHeaderStringValue = "(BLOCK_HEADER)"

	// BlockSignatureStringValue represents the string to be used to identify a block's signature
	BlockSignatureStringValue = "(SIGNATURE)"

	// BlockHeaderFinalInfoStringValue represents the string to be used to identify a block's header final info
	BlockHeaderFinalInfoStringValue = "(FINAL_INFO)"

	// BlockUnknownStringValue represents the string to be used to identify an unknown block
	BlockUnknownStringValue = "(UNKNOWN)"

	// BlockDefaultStringValue represents the message to identify a message that is undefined
	BlockDefaultStringValue = "Undefined message type"
)

// GetStringValue returns the string value of a given MessageType
func GetStringValue(msgType consensus.MessageType) string {
	switch msgType {
	case MtBlockBodyAndHeader:
		return BlockBodyAndHeaderStringValue
	case MtBlockBody:
		return BlockBodyStringValue
	case MtBlockHeader:
		return BlockHeaderStringValue
	case MtSignature:
		return BlockSignatureStringValue
	case MtBlockHeaderFinalInfo:
		return BlockHeaderFinalInfoStringValue
	case MtUnknown:
		return BlockUnknownStringValue
	default:
		return BlockDefaultStringValue
	}
}

// GetSubroundName returns the name of each Subround from a given Subround ID
func GetSubroundName(subroundId int) string {
	switch subroundId {
	case SrStartRound:
		return "(START_ROUND)"
	case SrBlock:
		return "(BLOCK)"
	case SrSignature:
		return "(SIGNATURE)"
	case SrEndRound:
		return "(END_ROUND)"
	default:
		return "Undefined subround"
	}
}
