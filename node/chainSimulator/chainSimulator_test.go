package chainSimulator

import (
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         core.OptionalUint64{},
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	chainSimulator.Close()
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:      api.NewNoApiInterface(),
		MinNodesPerShard:  1,
		MetaChainMinNodes: 1,
		InitialRound:      200000000,
		InitialEpoch:      100,
		InitialNonce:      100,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksAndEpochChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       100,
		MetaChainMinNodes:      100,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	facade, err := NewChainSimulatorFacade(chainSimulator)
	require.Nil(t, err)

	genesisBalances := make(map[string]*big.Int)
	for _, stakeWallet := range chainSimulator.initialWalletKeys.StakeWallets {
		initialAccount, errGet := facade.GetExistingAccountFromBech32AddressString(stakeWallet.Address.Bech32)
		require.Nil(t, errGet)

		genesisBalances[stakeWallet.Address.Bech32] = initialAccount.GetBalance()
	}

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(80)
	require.Nil(t, err)

	numAccountsWithIncreasedBalances := 0
	for _, stakeWallet := range chainSimulator.initialWalletKeys.StakeWallets {
		account, errGet := facade.GetExistingAccountFromBech32AddressString(stakeWallet.Address.Bech32)
		require.Nil(t, errGet)

		if account.GetBalance().Cmp(genesisBalances[stakeWallet.Address.Bech32]) > 0 {
			numAccountsWithIncreasedBalances++
		}
	}

	assert.True(t, numAccountsWithIncreasedBalances > 0)
}

func TestChainSimulator_SetState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	keyValueMap := map[string]string{
		"01": "01",
		"02": "02",
	}

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	err = chainSimulator.SetKeyValueForAddress(address, keyValueMap)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := chainSimulator.GetNodeHandler(0)
	keyValuePairs, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, keyValueMap, keyValuePairs)
}

func TestChainSimulator_SetEntireState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	balance := "431271308732096033771131"
	contractAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	accountState := &dtos.AddressState{
		Address:          contractAddress,
		Nonce:            new(uint64),
		Balance:          balance,
		Code:             "0061736d010000000129086000006000017f60027f7f017f60027f7f0060017f0060037f7f7f017f60037f7f7f0060017f017f0290020b03656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000103656e760b7369676e616c4572726f72000303656e76126d42756666657253746f726167654c6f6164000203656e76176d427566666572546f426967496e74556e7369676e6564000203656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e76136d42756666657253746f7261676553746f7265000203656e760f6d4275666665725365744279746573000503656e760e636865636b4e6f5061796d656e74000003656e7614626967496e7446696e697368556e7369676e6564000403656e7609626967496e744164640006030b0a010104070301000000000503010003060f027f0041a080080b7f0041a080080b074607066d656d6f7279020004696e697400110667657453756d00120361646400130863616c6c4261636b00140a5f5f646174615f656e6403000b5f5f686561705f6261736503010aca010a0e01017f4100100c2200100020000b1901017f419c8008419c800828020041016b220036020020000b1400100120004604400f0b4180800841191002000b16002000100c220010031a2000100c220010041a20000b1401017f100c2202200110051a2000200210061a0b1301017f100c220041998008410310071a20000b1401017f10084101100d100b210010102000100f0b0e0010084100100d1010100e10090b2201037f10084101100d100b210110102202100e220020002001100a20022000100f0b0300010b0b2f0200418080080b1c77726f6e67206e756d626572206f6620617267756d656e747373756d00419c80080b049cffffff",
		CodeHash:         "n9EviPlHS6EV+3Xp0YqP28T0IUfeAFRFBIRC1Jw6pyU=",
		RootHash:         "76cr5Jhn6HmBcDUMIzikEpqFgZxIrOzgNkTHNatXzC4=",
		CodeMetadata:     "BQY=",
		Owner:            "erd1ss6u80ruas2phpmr82r42xnkd6rxy40g9jl69frppl4qez9w2jpsqj8x97",
		DeveloperRewards: "5401004999998",
		Keys: map[string]string{
			"73756d": "0a",
		},
	}

	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{accountState})
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	nodeHandler := chainSimulator.GetNodeHandler(1)
	scAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(contractAddress)
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

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(contractAddress, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, accountState.Balance, account.Balance)
	require.Equal(t, accountState.DeveloperRewards, account.DeveloperReward)
	require.Equal(t, accountState.Code, account.Code)
	require.Equal(t, accountState.CodeHash, base64.StdEncoding.EncodeToString(account.CodeHash))
	require.Equal(t, accountState.CodeMetadata, base64.StdEncoding.EncodeToString(account.CodeMetadata))
	require.Equal(t, accountState.Owner, account.OwnerAddress)
	require.Equal(t, accountState.RootHash, base64.StdEncoding.EncodeToString(account.RootHash))
}

func TestChainSimulator_GetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	// the facade's GetAccount method requires that at least one block was produced over the genesis block
	_ = chainSimulator.GenerateBlocks(1)

	defer chainSimulator.Close()

	address := dtos.WalletAddress{
		Bech32: "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj",
	}
	address.Bytes, err = chainSimulator.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(address.Bech32)
	assert.Nil(t, err)

	account, err := chainSimulator.GetAccount(address)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), account.Nonce)
	assert.Equal(t, "0", account.Balance)

	nonce := uint64(37)
	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{
		{
			Address: address.Bech32,
			Nonce:   &nonce,
			Balance: big.NewInt(38).String(),
		},
	})
	assert.Nil(t, err)

	// without this call the test will fail because the latest produced block points to a state roothash that tells that
	// the account has the nonce 0
	_ = chainSimulator.GenerateBlocks(1)

	account, err = chainSimulator.GetAccount(address)
	assert.Nil(t, err)
	assert.Equal(t, uint64(37), account.Nonce)
	assert.Equal(t, "38", account.Balance)
}

func TestSimulator_SendTransactions(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(100))
	transferValue := big.NewInt(0).Mul(oneEgld, big.NewInt(5))

	wallet0, err := chainSimulator.GenerateAndMintWalletAddress(0, initialMinting)
	require.Nil(t, err)

	wallet1, err := chainSimulator.GenerateAndMintWalletAddress(1, initialMinting)
	require.Nil(t, err)

	wallet2, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	wallet3, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	wallet4, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	gasLimit := uint64(50000)
	tx0 := generateTransaction(wallet0.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx1 := generateTransaction(wallet1.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx3 := generateTransaction(wallet3.Bytes, 0, wallet4.Bytes, transferValue, "", gasLimit)

	maxNumOfBlockToGenerateWhenExecutingTx := 15

	t.Run("nil or empty slice of transactions should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(nil, 1)
		assert.Equal(t, errEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)

		sentTxs, errSend = chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(make([]*transaction.Transaction, 0), 1)
		assert.Equal(t, errEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("invalid max number of blocks to generate should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, 0)
		assert.Equal(t, errInvalidMaxNumOfBlocks, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("nil transaction in slice should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{nil}, 1)
		assert.ErrorIs(t, errSend, errNilTransaction)
		assert.Nil(t, sentTxs)
	})
	t.Run("2 transactions from different shard should call send correctly", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Equal(t, 2, len(sentTxs))
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet2)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(initialMinting, transferValue)
		expectedBalance.Add(expectedBalance, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
	t.Run("1 transaction should be sent correctly", func(t *testing.T) {
		_, errSend := chainSimulator.SendTxAndGenerateBlockTilTxIsExecuted(tx3, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet4)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(initialMinting, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
}

func generateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	minGasPrice := uint64(1000000000)
	txVersion := uint32(1)
	mockTxSignature := "sig"

	transferValue := big.NewInt(0).Set(value)
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     transferValue,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}
