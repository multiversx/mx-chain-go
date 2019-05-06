package consensus

// MessageType specifies what type of message was received
type MessageType int

// Message defines the data needed by spos to communicate between nodes over network in all subrounds
type Message struct {
	BlockHeaderHash []byte
	SubRoundData    []byte
	PubKey          []byte
	Signature       []byte
	MsgType         int
	TimeStamp       uint64
	RoundIndex      int32
}

// NewConsensusMessage creates a new Message object
func NewConsensusMessage(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg int,
	tms uint64,
	roundIndex int32,
) *Message {

	return &Message{
		BlockHeaderHash: blHeaderHash,
		SubRoundData:    subRoundData,
		PubKey:          pubKey,
		Signature:       sig,
		MsgType:         msg,
		TimeStamp:       tms,
		RoundIndex:      roundIndex,
	}
}
