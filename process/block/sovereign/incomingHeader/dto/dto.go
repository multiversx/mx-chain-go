package dto

import (
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

const (
	MinTopicsInTransferEvent = 5
	NumTransferTopics        = 3
	TokensIndex              = 2
)

const (
	EventIDExecutedOutGoingBridgeOp = "execute"
	EventIDDepositIncomingTransfer  = "deposit"
	EventIDChangeValidatorSet       = "changeValidatorSet"

	TopicIDConfirmedOutGoingOperation = "executedBridgeOp"
	TopicIDDepositIncomingTransfer    = "deposit"
)

// SCRInfo holds an incoming scr that is created based on an incoming cross chain event and its hash
type SCRInfo struct {
	SCR  *smartContractResult.SmartContractResult
	Hash []byte
}

// ConfirmedBridgeOp holds the hashes for a bridge operations that are confirmed from the main chain
type ConfirmedBridgeOp struct {
	HashOfHashes []byte
	Hash         []byte
}

// EventResult holds the result of processing an incoming cross chain event
type EventResult struct {
	SCR               *SCRInfo
	ConfirmedBridgeOp *ConfirmedBridgeOp
}

// EventsResult holds the results of processing incoming cross chain events
type EventsResult struct {
	Scrs               []*SCRInfo
	ConfirmedBridgeOps []*ConfirmedBridgeOp
}
