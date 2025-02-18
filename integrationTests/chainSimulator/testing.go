package chainSimulator

import (
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/errors"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CheckSetState -
func CheckSetState(t *testing.T, chainSimulator ChainSimulator, nodeHandler chainSimulatorProcess.NodeHandler) {
	keyValueMap := map[string]string{
		"01": "01",
		"02": "02",
	}

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	err := chainSimulator.SetKeyValueForAddress(address, keyValueMap)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	keyValuePairs, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, keyValueMap, keyValuePairs)
}

// CheckSetEntireState -
func CheckSetEntireState(t *testing.T, chainSimulator ChainSimulator, nodeHandler chainSimulatorProcess.NodeHandler, accountState *dtos.AddressState) {
	err := chainSimulator.SetStateMultiple([]*dtos.AddressState{accountState})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	scAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(accountState.Address)
	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)

	counterValue := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 10, int(counterValue))

	time.Sleep(time.Second)

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(accountState.Address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, accountState.Balance, account.Balance)
	require.Equal(t, accountState.DeveloperRewards, account.DeveloperReward)
	require.Equal(t, accountState.Code, account.Code)
	require.Equal(t, accountState.CodeHash, base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, accountState.CodeMetadata, base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, accountState.Owner, account.OwnerAddress)
	require.Equal(t, accountState.RootHash, base64.StdEncoding.EncodeToString(account.RootHash))
}

// CheckSetEntireStateWithRemoval -
func CheckSetEntireStateWithRemoval(t *testing.T, chainSimulator ChainSimulator, nodeHandler chainSimulatorProcess.NodeHandler, accountState *dtos.AddressState) {
	// activate the auto balancing tries so the results will be the same
	err := chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{accountState})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(2)
	require.Nil(t, err)

	scAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(accountState.Address)
	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)

	counterValue := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 10, int(counterValue))

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(accountState.Address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, accountState.Balance, account.Balance)
	require.Equal(t, accountState.DeveloperRewards, account.DeveloperReward)
	require.Equal(t, accountState.Code, account.Code)
	require.Equal(t, accountState.CodeHash, base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, accountState.CodeMetadata, base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, accountState.Owner, account.OwnerAddress)
	require.Equal(t, accountState.RootHash, base64.StdEncoding.EncodeToString(account.RootHash))

	// Now we remove the account
	err = chainSimulator.RemoveAccounts([]string{accountState.Address})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(2)
	require.Nil(t, err)

	account, _, err = nodeHandler.GetFacadeHandler().GetAccount(accountState.Address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, "0", account.Balance)
	require.Equal(t, "0", account.DeveloperReward)
	require.Equal(t, "", account.Code)
	require.Equal(t, "", base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, "", base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, "", account.OwnerAddress)
	require.Equal(t, "", base64.StdEncoding.EncodeToString(account.RootHash))

	// Set the state again
	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{accountState})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(2)
	require.Nil(t, err)

	account, _, err = nodeHandler.GetFacadeHandler().GetAccount(accountState.Address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)

	require.Equal(t, accountState.Balance, account.Balance)
	require.Equal(t, accountState.DeveloperRewards, account.DeveloperReward)
	require.Equal(t, accountState.Code, account.Code)
	require.Equal(t, accountState.CodeHash, base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, accountState.CodeMetadata, base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, accountState.Owner, account.OwnerAddress)
	require.Equal(t, accountState.RootHash, base64.StdEncoding.EncodeToString(account.RootHash))
}

// CheckGetAccount -
func CheckGetAccount(t *testing.T, chainSimulator ChainSimulator) {
	// the facade's GetAccount method requires that at least one block was produced over the genesis block
	err := chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	address := dtos.WalletAddress{
		Bech32: "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj",
	}
	address.Bytes, err = chainSimulator.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(address.Bech32)
	require.Nil(t, err)

	account, err := chainSimulator.GetAccount(address)
	require.Nil(t, err)
	require.Equal(t, uint64(0), account.Nonce)
	require.Equal(t, "0", account.Balance)

	nonce := uint64(37)
	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{
		{
			Address: address.Bech32,
			Nonce:   &nonce,
			Balance: big.NewInt(38).String(),
		},
	})
	require.Nil(t, err)

	// without this call the test will fail because the latest produced block points to a state roothash that tells that
	// the account has the nonce 0
	_ = chainSimulator.GenerateBlocks(1)

	account, err = chainSimulator.GetAccount(address)
	require.Nil(t, err)
	require.Equal(t, uint64(37), account.Nonce)
	require.Equal(t, "38", account.Balance)
}

// CheckGenerateTransactions -
func CheckGenerateTransactions(t *testing.T, chainSimulator ChainSimulator) {
	transferValue := big.NewInt(0).Mul(OneEGLD, big.NewInt(5))

	wallet0, err := chainSimulator.GenerateAndMintWalletAddress(0, InitialAmount)
	require.Nil(t, err)

	wallet1, err := chainSimulator.GenerateAndMintWalletAddress(1, InitialAmount)
	require.Nil(t, err)

	wallet2, err := chainSimulator.GenerateAndMintWalletAddress(2, InitialAmount)
	require.Nil(t, err)

	wallet3, err := chainSimulator.GenerateAndMintWalletAddress(2, InitialAmount)
	require.Nil(t, err)

	wallet4, err := chainSimulator.GenerateAndMintWalletAddress(2, InitialAmount)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	gasLimit := uint64(50000)
	tx0 := GenerateTransaction(wallet0.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx1 := GenerateTransaction(wallet1.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx3 := GenerateTransaction(wallet3.Bytes, 0, wallet4.Bytes, transferValue, "", gasLimit)

	maxNumOfBlockToGenerateWhenExecutingTx := 15

	t.Run("nil or empty slice of transactions should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(nil, 1)
		assert.Equal(t, errors.ErrEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)

		sentTxs, errSend = chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(make([]*transaction.Transaction, 0), 1)
		assert.Equal(t, errors.ErrEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("invalid max number of blocks to generate should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, 0)
		assert.Equal(t, errors.ErrInvalidMaxNumOfBlocks, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("nil transaction in slice should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{nil}, 1)
		assert.ErrorIs(t, errSend, errors.ErrNilTransaction)
		assert.Nil(t, sentTxs)
	})
	t.Run("2 transactions from different shard should call send correctly", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Equal(t, 2, len(sentTxs))
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet2)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(InitialAmount, transferValue)
		expectedBalance.Add(expectedBalance, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
	t.Run("1 transaction should be sent correctly", func(t *testing.T) {
		_, errSend := chainSimulator.SendTxAndGenerateBlockTilTxIsExecuted(tx3, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet4)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(InitialAmount, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
}
