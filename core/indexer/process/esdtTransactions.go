package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type esdtTransactionProc struct {
	esdtOperations map[string]struct{}
	argumentParser process.CallArgumentsParser
}

func newEsdtTransactionHandler() *esdtTransactionProc {
	esdtTxProc := &esdtTransactionProc{
		argumentParser: parsers.NewCallArgsParser(),
	}

	esdtTxProc.initESDTOperations()

	return esdtTxProc
}

func (etp *esdtTransactionProc) initESDTOperations() {
	etp.esdtOperations = map[string]struct{}{
		core.BuiltInFunctionESDTTransfer: {},
		core.BuiltInFunctionESDTBurn:     {},
		core.BuiltInFunctionESDTFreeze:   {},
		core.BuiltInFunctionESDTUnFreeze: {},
		core.BuiltInFunctionESDTWipe:     {},
		core.BuiltInFunctionESDTPause:    {},
		core.BuiltInFunctionESDTUnPause:  {},
	}
}

func (etp *esdtTransactionProc) isESDTTx(tx data.TransactionHandler) bool {
	txData := tx.GetData()

	function, _, err := etp.argumentParser.ParseData(string(txData))
	if err != nil {
		return false
	}

	_, ok := etp.esdtOperations[function]

	return ok
}

func (etp *esdtTransactionProc) getTokenIdentifierAndValue(
	tx data.TransactionHandler,
) (tokenIdentifier string, value string) {
	txData := tx.GetData()
	_, arguments, err := etp.argumentParser.ParseData(string(txData))
	if err != nil {
		return
	}

	if len(arguments) >= 1 {
		tokenIdentifier = string(arguments[0])
	}
	if len(arguments) >= 2 {
		bigValue := big.NewInt(0).SetBytes(arguments[1])

		value = bigValue.String()
	}

	return
}
