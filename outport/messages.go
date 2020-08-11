package outport

import "fmt"

// MessageKind is the kind of a message (that is passed from the Node to the outport sink)
type MessageKind uint32

const (
	ReservedFirstKind MessageKind = iota
	CommittedBlock
	ReservedLastKind
)

var messageKindNameByID = map[MessageKind]string{}

func init() {
	messageKindNameByID[ReservedFirstKind] = "ReservedFirstKind"
	messageKindNameByID[CommittedBlock] = "CommittedBlock"
	messageKindNameByID[ReservedLastKind] = "ReservedLastKind"
}

// MessageHandler is a message abstraction
type MessageHandler interface {
	GetSeqNum() uint32
	SetSeqNum(nonce uint32)
	GetKind() MessageKind
	SetKind(kind MessageKind)
	GetKindName() string
	DebugString() string
}

// Message is the implementation of the abstraction
type Message struct {
	SeqNum uint32
	Kind   MessageKind
}

// GetNonce gets the sequence number
func (message *Message) GetSeqNum() uint32 {
	return message.SeqNum
}

// SetNonce sets the sequence number
func (message *Message) SetSeqNum(seqNum uint32) {
	message.SeqNum = seqNum
}

// GetKind gets the message kind
func (message *Message) GetKind() MessageKind {
	return message.Kind
}

// SetKind sets the message kind
func (message *Message) SetKind(kind MessageKind) {
	message.Kind = kind
}

// GetKindName gets the kind name
func (message *Message) GetKindName() string {
	kindName := messageKindNameByID[message.Kind]
	return kindName
}

// DebugString is a debug representation of the message
func (message *Message) DebugString() string {
	kindName := messageKindNameByID[message.Kind]
	return fmt.Sprintf("[kind=%s num=%d]", kindName, message.SetSeqNum)
}

// MessageCommittedBlock represents a message
type MessageCommittedBlock struct {
	Message
}

// NewMessageCommittedBlock creates a message
func NewMessageCommittedBlock() *MessageCommittedBlock {
	message := &MessageCommittedBlock{}
	message.Kind = CommittedBlock
	return message
}

// //GetShardID() uint32
// 	GetNonce() uint64
// 	GetEpoch() uint32
// 	GetRound() uint64
// 	GetRootHash() []byte
// 	GetPrevHash() []byte
// 	GetSignature() []byte
// 	GetLeaderSignature() []byte
// 	GetChainID() []byte
// 	GetSoftwareVersion() []byte
// 	GetTimeStamp() uint64
// 	GetTxCount() uint32
// 	GetReceiptsHash() []byte
// 	GetAccumulatedFees() *big.Int
// 	GetDeveloperFees() *big.Int
// 	GetEpochStartMetaHash() []byte
