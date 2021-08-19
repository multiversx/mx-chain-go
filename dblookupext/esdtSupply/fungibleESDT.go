//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. supplyChanges.proto

package esdtSupply

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type fungibleESDTSupplyProc struct {
	marshalizer        marshal.Marshalizer
	storer             storage.Storer
	fungibleOperations map[string]struct{}
	mutex              sync.Mutex
}

func NewFungibleESDTSupplyProcessor(
	marshalizer marshal.Marshalizer,
	storer storage.Storer,
) (*fungibleESDTSupplyProc, error) {
	if check.IfNil(marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(storer) {
		return nil, core.ErrNilStore
	}

	return &fungibleESDTSupplyProc{
		fungibleOperations: map[string]struct{}{
			core.BuiltInFunctionESDTLocalBurn: {},
			core.BuiltInFunctionESDTLocalMint: {},
		},
	}, nil
}

func (fsp *fungibleESDTSupplyProc) ProcessLogs(blockHeaderHash []byte, logs map[string]*vmcommon.LogEntry) {
	fsp.mutex.Lock()
	defer fsp.mutex.Unlock()

	changes := make(map[string]*SupplyChange)
	for _, txLog := range logs {
		if fsp.shouldIgnoreLog(txLog) {
			continue
		}

		change, err := fsp.processLog(txLog)
		if err != nil {
			return
		}

		tokenIDStr := string(change.TokenID)
		_, found := changes[tokenIDStr]
		if !found {
			changes[tokenIDStr] = change
			continue
		}

		changes[tokenIDStr].Supply.Add(changes[tokenIDStr].Supply, change.Supply)
	}

	// calculate shard total supply
}

func (fsp *fungibleESDTSupplyProc) processLog(txLog *vmcommon.LogEntry) (*SupplyChange, error) {
	tokenIdentifier := txLog.Topics[0]
	bigValue := big.NewInt(0).SetBytes(txLog.Topics[2])

	if string(txLog.Identifier) == core.BuiltInFunctionESDTLocalBurn {
		bigValue = big.NewInt(0).Neg(bigValue)
	}

	supplyFromStorageBytes, err := fsp.storer.Get(tokenIdentifier)
	if err != nil {
		return &SupplyChange{
			TokenID: tokenIdentifier,
			Supply:  bigValue,
		}, nil
	}

	supplyFromStorage := &SupplyChange{}
	err = fsp.marshalizer.Unmarshal(supplyFromStorage, supplyFromStorageBytes)
	if err != nil {
		return nil, err
	}

	newSupply := big.NewInt(0).Add(supplyFromStorage.Supply, bigValue)
	return &SupplyChange{
		TokenID: tokenIdentifier,
		Supply:  newSupply,
	}, nil
}

func (fsp *fungibleESDTSupplyProc) shouldIgnoreLog(txLogs *vmcommon.LogEntry) bool {
	_, found := fsp.fungibleOperations[string(txLogs.Identifier)]

	return !found
}

func (fsp *fungibleESDTSupplyProc) RevertChanges(blockHash []byte) error {
	// revert changes in case of rollback
	return nil
}

func (fsp *fungibleESDTSupplyProc) OnNotarized() error {
	// delete changes from storage when block are notarized
	return nil
}
