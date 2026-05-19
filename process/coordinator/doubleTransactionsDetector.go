package coordinator

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const printReportHeaderNotCritical = "double transactions found (this is not critical, thus)\nshowing the whole block body:\n"
const printReportHeader = "double transactions found \nshowing the whole block body:\n"
const nilBlockBodyMessage = "nil block body in doubleTransactionsDetector.ProcessBlockBody"
const noDoubledTransactionsFoundMessage = "no double transactions found"
const doubledTransactionsFoundButFlagActive = "double transactions found but this is expected until the AddFailedRelayedTxToInvalidMBsDisableEpoch is deactivated"

// ArgsDoubleTransactionsDetector is the argument DTO structure used in the NewDoubleTransactionsDetector function
type ArgsDoubleTransactionsDetector struct {
	Marshaller          marshal.Marshalizer
	Hasher              hashing.Hasher
	EnableEpochsHandler common.EnableEpochsHandler
}

type doubleTransactionsDetector struct {
	marshaller          marshal.Marshalizer
	hasher              hashing.Hasher
	logger              logger.Logger
	enableEpochsHandler common.EnableEpochsHandler
}

// NewDoubleTransactionsDetector creates a new instance of doubleTransactionsDetector
func NewDoubleTransactionsDetector(args ArgsDoubleTransactionsDetector) (*doubleTransactionsDetector, error) {
	err := checkArgsDoubleTransactionsDetector(args)
	if err != nil {
		return nil, err
	}

	detector := &doubleTransactionsDetector{
		marshaller:          args.Marshaller,
		hasher:              args.Hasher,
		enableEpochsHandler: args.EnableEpochsHandler,
		logger:              log,
	}

	return detector, nil
}

func checkArgsDoubleTransactionsDetector(args ArgsDoubleTransactionsDetector) error {
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	return core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.AddFailedRelayedTxToInvalidMBsFlag,
	})
}

// ProcessBlockBody processes the block body provided in search of doubled transactions. If there are doubled transactions,
// this method will log as error the event providing as much information as possible
func (detector *doubleTransactionsDetector) ProcessBlockBody(body *block.Body) error {
	if body == nil {
		detector.logger.Error(nilBlockBodyMessage)
		return nil
	}

	transactions := make(map[string]int)
	doubleTransactionsExist := false
	printReport := strings.Builder{}

	for _, miniBlock := range body.MiniBlocks {
		mbHash, _ := core.CalculateHash(detector.marshaller, detector.hasher, miniBlock)
		log.Debug("checking for double transactions: miniblock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"type", miniBlock.Type,
			"num txs", len(miniBlock.TxHashes),
			"hash", mbHash)
		printReport.WriteString(fmt.Sprintf(" miniblock hash %s, type %s, %d -> %d\n",
			hex.EncodeToString(mbHash), miniBlock.Type.String(), miniBlock.SenderShardID, miniBlock.ReceiverShardID))

		for _, txHash := range miniBlock.TxHashes {
			transactions[string(txHash)]++
			printReport.WriteString(fmt.Sprintf("  tx hash %s\n", hex.EncodeToString(txHash)))

			doubleTransactionsExist = doubleTransactionsExist || transactions[string(txHash)] > 1
		}
	}

	if !doubleTransactionsExist {
		detector.logger.Debug(noDoubledTransactionsFoundMessage)
		return nil
	}

	relayedV1V2Disabled := detector.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsV1V2DisableFlag)

	if !relayedV1V2Disabled {
		if detector.enableEpochsHandler.IsFlagEnabled(common.AddFailedRelayedTxToInvalidMBsFlag) {
			detector.logger.Debug(doubledTransactionsFoundButFlagActive)
			return nil
		}

		detector.logger.Error(printReportHeaderNotCritical + printReport.String())
		return nil
	}

	detector.logger.Error(printReportHeader + printReport.String())
	return process.ErrDoubleTransactionsFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (detector *doubleTransactionsDetector) IsInterfaceNil() bool {
	return detector == nil
}
