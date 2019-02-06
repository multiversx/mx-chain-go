package spos

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// ConsensusData defines the data needed by spos to communicate between nodes over network in all subrounds
type ConsensusData struct {
	BlockHeaderHash []byte
	SubRoundData    []byte
	PubKey          []byte
	Signature       []byte
	MsgType         int
	TimeStamp       uint64
	RoundIndex      int32
}

// NewConsensusData creates a new ConsensusData object
func NewConsensusData(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg int,
	tms uint64,
	roundIndex int32,
) *ConsensusData {

	return &ConsensusData{
		BlockHeaderHash: blHeaderHash,
		SubRoundData:    subRoundData,
		PubKey:          pubKey,
		Signature:       sig,
		MsgType:         msg,
		TimeStamp:       tms,
		RoundIndex:      roundIndex,
	}
}

// Create method creates a new ConsensusData object
func (cnsdta *ConsensusData) Create() p2p.Creator {
	return &ConsensusData{}
}

// ID gets an unique id of the ConsensusData object
func (cnsdta *ConsensusData) ID() string {
	id := fmt.Sprintf("%d-%s-%d", cnsdta.RoundIndex, cnsdta.Signature, cnsdta.MsgType)
	return id
}
