package transactionAPI

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
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
	Parse(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData
}
