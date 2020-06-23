package iele

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

// DeployContract -
func DeployContract(
	tb testing.TB,
	senderAddressBytes []byte,
	senderNonce uint64,
	transferOnCalls *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCode string,
	txProc process.TransactionProcessor,
	accnts state.AccountsAdapter,
) error {

	//contract creation tx
	tx := vm.CreateTx(
		tb,
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	_, err := txProc.ProcessTransaction(tx)
	if err != nil {
		return err
	}

	_, err = accnts.Commit()
	if err != nil {
		return err
	}

	return nil
}

// CreateDeployTxData -
func CreateDeployTxData(scCode []byte) string {
	return strings.Join([]string{string(scCode), hex.EncodeToString(factory.IELEVirtualMachine), "0000"}, "@")
}
