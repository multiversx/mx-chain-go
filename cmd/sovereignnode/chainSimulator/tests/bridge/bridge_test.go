package bridge

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	esdtSafeWasmPath           = "../testdata/sov-esdt-safe.wasm"
	feeMarketWasmPath          = "../testdata/fee-market.wasm"
	issuePrice                 = "5000000000000000000"
)

// This test will:
// - deploy bridge contracts setup
// - issue a new fungible token
// - deposit some tokens in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsThenIssueAndDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	bridgeData := deploySovereignBridgeSetup(t, cs, wallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	nonce := GetNonce(t, nodeHandler, wallet.Bech32)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	depositAndCheckTokens(t, cs, wallet, nonce, bridgeData, tokenIdentifier, supply)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	checkOutGoingOperation(t, cs)
}

// This test will:
// - deploy bridge contracts setup
// - generate new wallet and set a token without prefix (a.k.a main chain token)
// - deposit the token in esdt-safe contract
// - check the sender balance is correct
// - check the token burned amount is correct after deposit
func TestSovereignChainSimulator_DeployBridgeContractsAndDepositMainChainToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outGoingSubscribedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents = []config.SubscribedEvent{
					{
						Identifier: "deposit",
						Addresses:  []string{outGoingSubscribedAddress},
					},
				}
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds = 1
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	mainChainToken := "MAIN-1a2b3c"

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)

	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	// set main chain token in system account
	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + mainChainToken))
	err = cs.SetKeyValueForAddress(chainSim.ESDTSystemAccount,
		map[string]string{
			tokenKey: "0400",
		},
	)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	expectedESDTSafeAddressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(outGoingSubscribedAddress)
	require.Nil(t, err)

	initialWallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeData := deploySovereignBridgeSetup(t, cs, initialWallet, esdtSafeWasmPath, feeMarketWasmPath)
	require.Equal(t, expectedESDTSafeAddressBytes, bridgeData.ESDTSafeAddress)

	// generate wallet and set main chain token supply
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	mainChainTokenSupply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, mainChainToken, 0, esdt.ESDigitalToken{Value: mainChainTokenSupply})
	nonce := GetNonce(t, nodeHandler, wallet.Bech32)

	depositAndCheckTokens(t, cs, wallet, nonce, bridgeData, mainChainToken, mainChainTokenSupply)

	// Wait for outgoing operations to get unconfirmed and check we have one, which is also saved in storage
	time.Sleep(time.Second)

	checkOutGoingOperation(t, cs)
}

func depositAndCheckTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce uint64,
	bridgeData ArgsBridgeSetup,
	mainChainToken string,
	mainChainTokenSupply *big.Int,
) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: mainChainToken,
		Nonce:      0,
		Amount:     amountToDeposit,
	})

	txResult := Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(0).Sub(mainChainTokenSupply, amountToDeposit).String(), tokens[mainChainToken].GetValue().String())

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(mainChainToken)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)
}

func checkOutGoingOperation(t *testing.T, cs chainSim.ChainSimulator) {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	outGoingOps := nodeHandler.GetRunTypeComponents().OutGoingOperationsPoolHandler().GetUnconfirmedOperations()
	require.Len(t, outGoingOps, 1)
	require.Len(t, outGoingOps[0].OutGoingOperations, 1)
	outGoingOp := outGoingOps[0].OutGoingOperations[0]
	savedMarshalledTx, err := nodeHandler.GetDataComponents().StorageService().Get(dataRetriever.TransactionUnit, outGoingOp.Hash)
	require.Nil(t, err)
	require.NotNil(t, savedMarshalledTx)

	savedTx := &transaction.Transaction{}
	err = nodeHandler.GetCoreComponents().InternalMarshalizer().Unmarshal(savedTx, savedMarshalledTx)
	require.Nil(t, err)

	expectedSavedTx := &transaction.Transaction{
		GasPrice: nodeHandler.GetCoreComponents().EconomicsData().MinGasPrice(),
		GasLimit: nodeHandler.GetCoreComponents().EconomicsData().ComputeGasLimit(
			&transaction.Transaction{
				Data: outGoingOp.Data,
			}),
		Data: outGoingOp.Data,
	}
	require.Equal(t, expectedSavedTx, savedTx)

	// Generate extra blocks after outgoing operations are created
	err = cs.GenerateBlocks(10)
	require.Nil(t, err)
}
