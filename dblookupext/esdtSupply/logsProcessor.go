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

func (lp *logsProcessor) processLogs(blockNonce uint64, logs map[string]data.LogHandler, isRevert bool) error {
	shouldProcess, err := lp.nonceProc.shouldProcessLog(blockNonce, isRevert)
	if err != nil {
		return err
	}
	if !shouldProcess {
		return nil
	}

	supplies := make(map[string]*SupplyESDT)
	for _, logHandler := range logs {
		if check.IfNil(logHandler) {
			continue
		}

		errProc := lp.processLog(logHandler, supplies, isRevert)
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

	bigValue := big.NewInt(0).SetBytes(txLog.Topics[2])

	negValue := string(txLog.Identifier) == core.BuiltInFunctionESDTLocalBurn || string(txLog.Identifier) == core.BuiltInFunctionESDTNFTBurn ||
		string(txLog.Identifier) == core.BuiltInFunctionESDTWipe
	xorBooleanVariable := negValue != isRevert
	// need this because
	//  negValue | isRevert  => res
	//   false   |   false   => false
	//   false   |   true    => true
	//   true    |   true    => false
	//   true    |   false   => true
	if xorBooleanVariable {
		bigValue = big.NewInt(0).Neg(bigValue)
	}

	tokenIDStr := string(tokenIdentifier)
	_, found := supplies[tokenIDStr]
	if found {
		supplies[tokenIDStr].Supply.Add(supplies[tokenIDStr].Supply, bigValue)
		return nil
	}

	supplyFromStorage, err := lp.getSupply(tokenIdentifier)
	if err != nil {
		return err
	}

	newSupply := big.NewInt(0).Add(supplyFromStorage.Supply, bigValue)
	supplies[tokenIDStr] = &SupplyESDT{
		Supply: newSupply,
	}

	return nil
}

func (lp *logsProcessor) getSupply(tokenIdentifier []byte) (*SupplyESDT, error) {
	supplyFromStorageBytes, err := lp.suppliesStorer.Get(tokenIdentifier)
	if err != nil {
		return &SupplyESDT{
			big.NewInt(0),
		}, nil
	}

	supplyFromStorage := &SupplyESDT{}
	err = lp.marshalizer.Unmarshal(supplyFromStorage, supplyFromStorageBytes)
	if err != nil {
		return nil, err
	}

	return supplyFromStorage, nil
}

func (lp *logsProcessor) shouldIgnoreEvent(event *transaction.Event) bool {
	_, found := lp.fungibleOperations[string(event.Identifier)]

	return !found
}

func (lp *logsProcessor) getESDTSupply(token string) (string, error) {
	supplyBytes, err := lp.suppliesStorer.Get([]byte(token))
	if err != nil && err == storage.ErrKeyNotFound {
		return big.NewInt(0).String(), nil
	}
	if err != nil {
		return "", err
	}

	supply := &SupplyESDT{}
	err = lp.marshalizer.Unmarshal(supply, supplyBytes)
	if err != nil {
		return "", err
	}

	return supply.Supply.String(), nil
}
