package smartContract

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/data"
)

// GetLatestTestError should only be used in tests!
// It locates the latest error in the collection of smart contracts results
func GetLatestTestError(scProcessorAsInterface interface{}) error {
	scProcessor, ok := scProcessorAsInterface.(*scProcessor)
	if !ok {
		return nil
	}

	scrProvider, ok := scProcessor.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return nil
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

		returnCodeHex := parts[0]
		returnCode, err := hex.DecodeString(returnCodeHex)
		if err == nil {
			returnCodeAsString := string(returnCode)
			if returnCodeAsString == "ok" || returnCodeAsString == "" {
				return nil
			}
		}

		return fmt.Errorf(returnCodeHex)
	}

	return nil
}

// GetAllSCRs returns all generated scrs
func GetAllSCRs(scProcessorAsInterface interface{}) []data.TransactionHandler {
	scProc, ok := scProcessorAsInterface.(*scProcessor)
	if !ok {
		return nil
	}

	scrProvider, ok := scProc.scrForwarder.(interface {
		GetIntermediateTransactions() []data.TransactionHandler
	})
	if !ok {
		return nil
	}

	return scrProvider.GetIntermediateTransactions()
}
