package smartContract

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

// TestScProcessor extends scProcessor and is used in tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestScProcessor struct {
	*scProcessor
}

// NewTestScProcessor -
func NewTestScProcessor(internalData *scProcessor) *TestScProcessor {
	return &TestScProcessor{internalData}
}

// GetCompositeTestError composes all errors found in the logs or by parsing the scr forwarder's contents
func (tsp *TestScProcessor) GetCompositeTestError() error {
	var returnError error

	currentEpoch := tsp.enableEpochsHandler.GetCurrentEpoch()
	if tsp.enableEpochsHandler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(currentEpoch) {
		allLogs := tsp.txLogsProcessor.GetAllCurrentLogs()
		for _, logs := range allLogs {
			for _, event := range logs.GetLogEvents() {
				if string(event.GetIdentifier()) == core.SignalErrorOperation {
					returnError = wrapErrorIfNotContains(returnError, string(event.GetTopics()[1]))
				}
			}
		}
	}

	scrProvider, ok := tsp.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return returnError
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
				return returnError
			}
			return wrapErrorIfNotContains(returnError, returnCodeAsString)
		}

		return wrapErrorIfNotContains(returnError, returnCodeHex)
	}

	tsp.txLogsProcessor.Clean()
	tsp.scrForwarder.CreateBlockStarted()

	return returnError
}

func wrapErrorIfNotContains(originalError error, msg string) error {
	if originalError == nil {
		return fmt.Errorf(msg)
	}

	alreadyContainsMessage := strings.Contains(originalError.Error(), msg)
	if alreadyContainsMessage {
		return originalError
	}

	return fmt.Errorf("%s: %s", originalError.Error(), msg)
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
