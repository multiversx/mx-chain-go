//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. processedBlockNonce.proto

package esdtSupply

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	storageRepo "github.com/ElrondNetwork/elrond-go-storage"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type logsProcessor struct {
	marshalizer        marshal.Marshalizer
	suppliesStorer     storage.Storer
	nonceProc          *nonceProcessor
	fungibleOperations map[string]struct{}
}

func newLogsProcessor(
	marshalizer marshal.Marshalizer,
	suppliesStorer storage.Storer,
) *logsProcessor {
	nonceProc := newNonceProcessor(marshalizer, suppliesStorer)

	return &logsProcessor{
		nonceProc:      nonceProc,
		marshalizer:    marshalizer,
		suppliesStorer: suppliesStorer,
		fungibleOperations: map[string]struct{}{
			core.BuiltInFunctionESDTLocalBurn:      {},
			core.BuiltInFunctionESDTLocalMint:      {},
			core.BuiltInFunctionESDTWipe:           {},
			core.BuiltInFunctionESDTNFTCreate:      {},
			core.BuiltInFunctionESDTNFTAddQuantity: {},
			core.BuiltInFunctionESDTNFTBurn:        {},
		},
	}
}

func (lp *logsProcessor) processLogs(blockNonce uint64, logs map[string]*data.LogData, isRevert bool) error {
	shouldProcess, err := lp.nonceProc.shouldProcessLog(blockNonce, isRevert)
	if err != nil {
		return err
	}
	if !shouldProcess {
		return nil
	}

	supplies := make(map[string]*SupplyESDT)
	for _, logHandler := range logs {
		if logHandler == nil || check.IfNil(logHandler.LogHandler) {
			continue
		}

		errProc := lp.processLog(logHandler.LogHandler, supplies, isRevert)
		if errProc != nil {
			return errProc
		}
	}

	err = lp.saveSupplies(supplies)
	if err != nil {
		return err
	}

	return lp.nonceProc.saveNonceInStorage(blockNonce)
}

func (lp *logsProcessor) processLog(txLog data.LogHandler, supplies map[string]*SupplyESDT, isRevert bool) error {
	for _, entryHandler := range txLog.GetLogEvents() {
		if check.IfNil(entryHandler) {
			continue
		}

		event, ok := entryHandler.(*transaction.Event)
		if !ok {
			continue
		}

		if lp.shouldIgnoreEvent(event) {
			continue
		}

		err := lp.processEvent(event, supplies, isRevert)
		if err != nil {
			return err
		}
	}

	return nil
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

func (lp *logsProcessor) processEvent(txLog *transaction.Event, supplies map[string]*SupplyESDT, isRevert bool) error {
	if len(txLog.Topics) < 3 {
		return nil
	}

	tokenIdentifier := txLog.Topics[0]
	if len(txLog.Topics[1]) != 0 {
		nonceBytes := txLog.Topics[1]
		nonceHexStr := hex.EncodeToString(nonceBytes)

		tokenIdentifier = bytes.Join([][]byte{tokenIdentifier, []byte(nonceHexStr)}, []byte("-"))
	}

	valueFromEvent := big.NewInt(0).SetBytes(txLog.Topics[2])

	tokenIDStr := string(tokenIdentifier)
	tokenSupply, found := supplies[tokenIDStr]
	if found {
		lp.updateTokenSupply(tokenSupply, valueFromEvent, string(txLog.Identifier), isRevert)
		return nil
	}

	supply, err := lp.getESDTSupply(tokenIdentifier)
	if err != nil {
		return err
	}

	supplies[tokenIDStr] = supply
	lp.updateTokenSupply(supplies[tokenIDStr], valueFromEvent, string(txLog.Identifier), isRevert)

	return nil
}

func (lp *logsProcessor) updateTokenSupply(tokenSupply *SupplyESDT, valueFromEvent *big.Int, eventIdentifier string, isRevert bool) {
	isBurnOp := eventIdentifier == core.BuiltInFunctionESDTLocalBurn || eventIdentifier == core.BuiltInFunctionESDTNFTBurn ||
		eventIdentifier == core.BuiltInFunctionESDTWipe
	isMintOp := eventIdentifier == core.BuiltInFunctionESDTNFTAddQuantity || eventIdentifier == core.BuiltInFunctionESDTLocalMint ||
		eventIdentifier == core.BuiltInFunctionESDTNFTCreate

	negativeValueFromEvent := big.NewInt(0).Neg(valueFromEvent)

	switch {
	case isMintOp && !isRevert:
		// normal processing mint - add to supply and add to minted
		tokenSupply.Minted.Add(tokenSupply.Minted, valueFromEvent)
		tokenSupply.Supply.Add(tokenSupply.Supply, valueFromEvent)
	case isMintOp && isRevert:
		// reverted mint - subtract from supply and subtract from minted
		tokenSupply.Minted.Add(tokenSupply.Minted, negativeValueFromEvent)
		tokenSupply.Supply.Add(tokenSupply.Supply, negativeValueFromEvent)
	case isBurnOp && !isRevert:
		// normal processing burn - subtract from supply and add to burn
		tokenSupply.Burned.Add(tokenSupply.Burned, valueFromEvent)
		tokenSupply.Supply.Add(tokenSupply.Supply, negativeValueFromEvent)
	case isBurnOp && isRevert:
		// reverted burn - subtract from burned and add to supply
		tokenSupply.Burned.Add(tokenSupply.Burned, negativeValueFromEvent)
		tokenSupply.Supply.Add(tokenSupply.Supply, valueFromEvent)
	}
}

func (lp *logsProcessor) getESDTSupply(tokenIdentifier []byte) (*SupplyESDT, error) {
	supplyFromStorageBytes, err := lp.suppliesStorer.Get(tokenIdentifier)
	if err != nil {
		if err == storageRepo.ErrKeyNotFound {
			return newSupplyESDTZero(), nil
		}

		return nil, err
	}

	supplyFromStorage := &SupplyESDT{}
	err = lp.marshalizer.Unmarshal(supplyFromStorage, supplyFromStorageBytes)
	if err != nil {
		return nil, err
	}

	makePropertiesNotNil(supplyFromStorage)
	return supplyFromStorage, nil
}

func (lp *logsProcessor) shouldIgnoreEvent(event *transaction.Event) bool {
	_, found := lp.fungibleOperations[string(event.Identifier)]

	return !found
}

func newSupplyESDTZero() *SupplyESDT {
	return &SupplyESDT{
		Burned: big.NewInt(0),
		Minted: big.NewInt(0),
		Supply: big.NewInt(0),
	}
}

func makePropertiesNotNil(supplyESDT *SupplyESDT) {
	if supplyESDT.Supply == nil {
		supplyESDT.Supply = big.NewInt(0)
	}
	if supplyESDT.Minted == nil {
		supplyESDT.Minted = big.NewInt(0)
	}
	if supplyESDT.Burned == nil {
		supplyESDT.Burned = big.NewInt(0)
	}
}
