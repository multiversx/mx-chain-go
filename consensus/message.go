//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. message.proto
package consensus

import "github.com/multiversx/mx-chain-core-go/core"

// MessageType specifies what type of message was received
type MessageType int

// NewConsensusMessage creates a new Message object
func NewConsensusMessage(
	headerHash []byte,
	signatureShare []byte,
	body []byte,
	header []byte,
	pubKey []byte,
	signature []byte,
	msgType int,
	roundIndex int64,
	chainID []byte,
	pubKeysBitmap []byte,
	aggregateSignature []byte,
	leaderSignature []byte,
	originatorPid core.PeerID,
	invalidSigners []byte,
	processedHeaderHash []byte,
) *Message {
	return &Message{
		HeaderHash:          headerHash,
		SignatureShare:      signatureShare,
		Body:                body,
		Header:              header,
		PubKey:              pubKey,
		Signature:           signature,
		MsgType:             int64(msgType),
		RoundIndex:          roundIndex,
		ChainID:             chainID,
		PubKeysBitmap:       pubKeysBitmap,
		AggregateSignature:  aggregateSignature,
		LeaderSignature:     leaderSignature,
		OriginatorPid:       originatorPid.Bytes(),
		InvalidSigners:      invalidSigners,
		ProcessedHeaderHash: processedHeaderHash,
	}
}
