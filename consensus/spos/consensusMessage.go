package spos

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
