package smartContract

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
)

// TestScProcessor extends scProcessor and is used in tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestScProcessor struct {
	*scProcessor
}

func NewTestScProcessor(internalData *scProcessor) *TestScProcessor {
	return &TestScProcessor{internalData}
}

// GetLatestTestError should only be used in tests!
// It locates the latest error in the collection of smart contracts results
// TODO remove this file as it is a horrible hack to test some conditions
func (tsp *TestScProcessor) GetLatestTestError() error {

	scrProvider, ok := tsp.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return nil
	}

	scResults := scrProvider.GetIntermediateTransactions()

	for i := len(scResults) - 1; i >= 0; i-- {
		_, isScr := scResults[i].(*smartContractResult.SmartContractResult)
		if !isScr {
			continue
		}

		tx := scResults[i]
		txData := string(tx.GetData())
		dataTrimmed := strings.Trim(txData, "@")

		parts := strings.Split(dataTrimmed, "@")
		if len(parts) == 0 {
			continue
		}

		returnCodeHex := parts[0]
		returnCode, err := hex.DecodeString(returnCodeHex)
		if err == nil {
			returnCodeAsString := string(returnCode)
			if returnCodeAsString == "ok" || returnCodeAsString == "" {
				return nil
			}
			return fmt.Errorf(returnCodeAsString)
		}

		return fmt.Errorf(returnCodeHex)
	}

	return nil
}

// GetGasRemaining returns the remaining gas from the last transaction
func (tsp *TestScProcessor) GetGasRemaining() uint64 {
	scrProvider, ok := tsp.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return 0
	}

	scResults := scrProvider.GetIntermediateTransactions()
	for i := len(scResults) - 1; i >= 0; i-- {
		tx := scResults[i]
		txData := string(tx.GetData())
		dataTrimmed := strings.Trim(txData, "@")

		parts := strings.Split(dataTrimmed, "@")
		if len(parts) == 0 {
			continue
		}

		return tx.GetGasLimit()
	}

	return 0
}

// GetAllSCRs returns all generated scrs
func (tsp *TestScProcessor) GetAllSCRs() []data.TransactionHandler {
	scrProvider, ok := tsp.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return nil
	}

	return scrProvider.GetIntermediateTransactions()
}

// CleanGasRefunded cleans the gas computation handler
func (tsp *TestScProcessor) CleanGasRefunded() {
	tsp.gasHandler.Init()
}
