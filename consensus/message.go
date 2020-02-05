package consensus

// MessageType specifies what type of message was received
type MessageType int

// Message defines the data needed by spos to communicate between nodes over network in all subrounds
type Message struct {
	BlockHeaderHash    []byte
	SubRoundData       []byte
	PubKey             []byte
	Signature          []byte
	MsgType            int
	RoundIndex         int64
	ChainID            []byte
	PubKeysBitmap      []byte
	AggregateSignature []byte
	LeaderSignature    []byte
}

// NewConsensusMessage creates a new Message object
func NewConsensusMessage(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg int,
	roundIndex int64,
	chainID []byte,
	pubKeysBitmap []byte,
	aggregateSignature []byte,
	leaderSignature []byte,
) *Message {
	return &Message{
		BlockHeaderHash:    blHeaderHash,
		SubRoundData:       subRoundData,
		PubKey:             pubKey,
		Signature:          sig,
		MsgType:            msg,
		RoundIndex:         roundIndex,
		ChainID:            chainID,
		PubKeysBitmap:      pubKeysBitmap,
		AggregateSignature: aggregateSignature,
		LeaderSignature:    leaderSignature,
	}
}
