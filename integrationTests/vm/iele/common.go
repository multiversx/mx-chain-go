package iele

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
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
) {

	//contract creation tx
	tx := vm.CreateTx(
		tb,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	err := txProc.ProcessTransaction(tx)
	assert.Nil(tb, err)

	_, err = accnts.Commit()
	assert.Nil(tb, err)
}
