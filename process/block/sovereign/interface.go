package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

type OutgoingOperationsFormatter interface {
	CreateOutgoingTxData(logs []*data.LogData) []byte
	IsInterfaceNil() bool
}
