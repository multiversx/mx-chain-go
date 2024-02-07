package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

// OutgoingOperationsFormatter collects relevant outgoing events for bridge from the logs and creates outgoing data
// that needs to be signed by validators to bridge tokens
type OutgoingOperationsFormatter interface {
	CreateOutgoingTxsData(logs []*data.LogData) [][]byte
	IsInterfaceNil() bool
}

// RoundHandler should be able to provide the current round
type RoundHandler interface {
	Index() int64
	IsInterfaceNil() bool
}
