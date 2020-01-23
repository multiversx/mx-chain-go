//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. message.proto
package consensus

// MessageType specifies what type of message was received
type MessageType int

// NewConsensusMessage creates a new Message object
func NewConsensusMessage(
	blHeaderHash []byte,
	subRoundData []byte,
	pubKey []byte,
	sig []byte,
	msg int,
	tms uint64,
	roundIndex int64,
) *Message {

	return &Message{
		BlockHeaderHash: blHeaderHash,
		SubRoundData:    subRoundData,
		PubKey:          pubKey,
		Signature:       sig,
		MsgType:         int64(msg),
		TimeStamp:       tms,
		RoundIndex:      roundIndex,
	}
}
