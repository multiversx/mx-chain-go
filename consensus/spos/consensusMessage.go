package spos

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// ConsensusMessage defines the data needed by spos to communicate between nodes over network in all subrounds
type ConsensusMessage struct {
	BlockHeaderHash []byte
	SubRoundData    []byte
	PubKey          []byte
	Signature       []byte
	MsgType         int
	TimeStamp       uint64
	RoundIndex      int32
}

// NewConsensusMessage creates a new ConsensusMessage object
func NewConsensusMessage(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg int,
	tms uint64,
	roundIndex int32,
) *ConsensusMessage {

	return &ConsensusMessage{
		BlockHeaderHash: blHeaderHash,
		SubRoundData:    subRoundData,
		PubKey:          pubKey,
		Signature:       sig,
		MsgType:         msg,
		TimeStamp:       tms,
		RoundIndex:      roundIndex,
	}
}

// Create method creates a new ConsensusMessage object
func (cnsdta *ConsensusMessage) Create() p2p.Creator {
	return &ConsensusMessage{}
}

// ID gets an unique id of the ConsensusMessage object
func (cnsdta *ConsensusMessage) ID() string {
	id := fmt.Sprintf("%d-%s-%d", cnsdta.RoundIndex, cnsdta.Signature, cnsdta.MsgType)
	return id
}
