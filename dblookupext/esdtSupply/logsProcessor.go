package esdtSupply

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type logsProcessor struct {
	marshalizer        marshal.Marshalizer
	suppliesStorer     storage.Storer
	fungibleOperations map[string]struct{}
}

func newLogsProcessor(
	marshalizer marshal.Marshalizer,
	suppliesStorer storage.Storer,
) *logsProcessor {
	return &logsProcessor{
		marshalizer:    marshalizer,
		suppliesStorer: suppliesStorer,
		fungibleOperations: map[string]struct{}{
			core.BuiltInFunctionESDTLocalBurn:      {},
			core.BuiltInFunctionESDTLocalMint:      {},
			core.BuiltInFunctionESDTNFTCreate:      {},
			core.BuiltInFunctionESDTNFTAddQuantity: {},
			core.BuiltInFunctionESDTNFTBurn:        {},
		},
	}
}

func (lp *logsProcessor) processLogs(logs map[string]*vmcommon.LogEntry, isRevert bool) error {
	supplies := make(map[string]*SupplyESDT)
	for _, txLog := range logs {
		if lp.shouldIgnoreLog(txLog) {
			continue
		}

		tokenSupply, tokenIdentifier, err := lp.processLog(txLog, isRevert)
		if err != nil {
			return err
		}

		tokenIDStr := string(tokenIdentifier)
		_, found := supplies[tokenIDStr]
		if !found {
			supplies[tokenIDStr] = tokenSupply
			continue
		}

		supplies[tokenIDStr].Supply.Add(supplies[tokenIDStr].Supply, tokenSupply.Supply)
	}

	return lp.saveSupplies(supplies)
}

func (lp *logsProcessor) saveSupplies(supplies map[string]*SupplyESDT) error {
	for identifier, supplyESDT := range supplies {
		supplyESDTBytes, err := lp.marshalizer.Marshal(supplyESDT)
		if err != nil {
			return err
		}

		err = lp.suppliesStorer.Put([]byte(identifier), supplyESDTBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (lp *logsProcessor) processLog(txLog *vmcommon.LogEntry, isRevert bool) (*SupplyESDT, []byte, error) {
	tokenIdentifier := txLog.Topics[0]
	if len(txLog.Topics[1]) != 0 {
		tokenIdentifier = bytes.Join([][]byte{tokenIdentifier, txLog.Topics[1]}, []byte("-"))
	}

	bigValue := big.NewInt(0).SetBytes(txLog.Topics[2])

	negValue := string(txLog.Identifier) == core.BuiltInFunctionESDTLocalBurn || string(txLog.Identifier) == core.BuiltInFunctionESDTNFTBurn
	shouldNegValue := negValue && !isRevert
	if shouldNegValue {
		bigValue = big.NewInt(0).Neg(bigValue)
	}

	supplyFromStorageBytes, err := lp.suppliesStorer.Get(tokenIdentifier)
	if err != nil {
		return &SupplyESDT{
			Supply: bigValue,
		}, tokenIdentifier, nil
	}

	supplyFromStorage := &SupplyESDT{}
	err = lp.marshalizer.Unmarshal(supplyFromStorage, supplyFromStorageBytes)
	if err != nil {
		return nil, nil, err
	}

	newSupply := big.NewInt(0).Add(supplyFromStorage.Supply, bigValue)
	return &SupplyESDT{
		Supply: newSupply,
	}, tokenIdentifier, nil
}

func (lp *logsProcessor) shouldIgnoreLog(txLogs *vmcommon.LogEntry) bool {
	_, found := lp.fungibleOperations[string(txLogs.Identifier)]

	return !found
}
