package testscommon

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const noDoubledTransactionsFoundMessage = "no double transactions found"
const printReportHeader = "double transactions found (this is not critical, thus)\nshowing the whole block body:\n"

var log = logger.GetOrCreate("testcommon")

// PanicDoubleTransactionsDetector -
type PanicDoubleTransactionsDetector struct {
}

// ProcessBlockBody -
func (detector *PanicDoubleTransactionsDetector) ProcessBlockBody(body *block.Body) {
	transactions := make(map[string]int)
	doubleTransactionsExist := false
	printReport := strings.Builder{}

	for _, miniBlock := range body.MiniBlocks {
		log.Debug("checking for double transactions: miniblock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"type", miniBlock.Type,
			"num txs", len(miniBlock.TxHashes))
		printReport.WriteString(fmt.Sprintf(" miniblock type %s, %d -> %d\n",
			miniBlock.Type.String(), miniBlock.SenderShardID, miniBlock.ReceiverShardID))

		for _, txHash := range miniBlock.TxHashes {
			transactions[string(txHash)]++
			printReport.WriteString(fmt.Sprintf("  tx hash %s\n", hex.EncodeToString(txHash)))

			doubleTransactionsExist = doubleTransactionsExist || transactions[string(txHash)] > 1
		}
	}

	if !doubleTransactionsExist {
		log.Debug(noDoubledTransactionsFoundMessage)
		return
	}

	log.Error(printReportHeader + printReport.String())
	panic("PanicDoubleTransactionsDetector.ProcessBlockBody - test failed")
}

// IsInterfaceNil -
func (detector *PanicDoubleTransactionsDetector) IsInterfaceNil() bool {
	return detector == nil
}
