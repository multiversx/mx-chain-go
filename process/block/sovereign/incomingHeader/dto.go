package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

const (
	minTopicsInTransferEvent  = 5
	numTransferTopics         = 3
	numExecutedBridgeOpTopics = 3
	tokensIndex               = 2
	hashOfHashesIndex         = 1
	hashOfOperationIndex      = 2
)

const (
	eventIDExecutedOutGoingBridgeOp = "execute"
	eventIDDepositIncomingTransfer  = "deposit"

	topicIDConfirmedOutGoingOperation = "executedBridgeOp"
	topicIDDepositIncomingTransfer    = "deposit"
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
