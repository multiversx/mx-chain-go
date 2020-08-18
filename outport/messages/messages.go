package messages

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
)

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

// GetSeqNum gets the sequence number
func (message *Message) GetSeqNum() uint32 {
	return message.SeqNum
}

// SetSeqNum sets the sequence number
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
	return fmt.Sprintf("[kind=%s num=%d]", kindName, message.GetSeqNum())
}

// MessageCommittedBlock represents a message
type MessageCommittedBlock struct {
	Message
	Header *CommittedBlockHeader

	RegularTransactions  *marshaling.SerializableMapStringTransactionHandler
	SmartContractResults *marshaling.SerializableMapStringTransactionHandler
	RewardTransactions   *marshaling.SerializableMapStringTransactionHandler
	InvalidTransactions  *marshaling.SerializableMapStringTransactionHandler
	Receipts             *marshaling.SerializableMapStringTransactionHandler
	SmartContractLogs    []data.LogHandler
}

// NewMessageCommittedBlock creates a message
func NewMessageCommittedBlock(header data.HeaderHandler) *MessageCommittedBlock {
	message := &MessageCommittedBlock{}
	message.Kind = CommittedBlock
	message.Header = NewCommittedBlockHeader(header)
	return message
}

// CommittedBlockHeader represents a block header
type CommittedBlockHeader struct {
	Shard           uint32
	Nonce           uint64
	Epoch           uint32
	Round           uint64
	RootHash        string
	PreviousHash    string
	Chain           string
	SoftwareVersion string
	Timestamp       uint64
	TxCount         uint32
	AccumulatedFees string
	DeveloperFees   string
}

// NewCommittedBlockHeader creates a CommittedBlockHeader
func NewCommittedBlockHeader(header data.HeaderHandler) *CommittedBlockHeader {
	return &CommittedBlockHeader{
		Shard:           header.GetShardID(),
		Nonce:           header.GetNonce(),
		Round:           header.GetRound(),
		RootHash:        hex.EncodeToString(header.GetRootHash()),
		PreviousHash:    hex.EncodeToString(header.GetPrevHash()),
		Chain:           string(header.GetChainID()),
		SoftwareVersion: string(header.GetSoftwareVersion()),
		Timestamp:       header.GetTimeStamp(),
		TxCount:         header.GetTxCount(),
		AccumulatedFees: header.GetAccumulatedFees().String(),
		DeveloperFees:   header.GetDeveloperFees().String(),
	}
}
