package transactionAPI

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

type feeComputer interface {
	ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int
	IsInterfaceNil() bool
}

// LogsFacade defines the interface of a logs facade
type LogsFacade interface {
	GetLog(logKey []byte, epoch uint32) (*transaction.ApiLogs, error)
	IsInterfaceNil() bool
}

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}
