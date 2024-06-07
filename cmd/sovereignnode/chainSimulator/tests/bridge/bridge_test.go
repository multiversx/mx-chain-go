package bridge

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	esdtSafeWasmPath           = "../testdata/esdt-safe.wasm"
	feeMarketWasmPath          = "../testdata/fee-market.wasm"
	issuePrice                 = "5000000000000000000"
)

var log = logger.GetOrCreate("dsa")

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

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.GeneralConfig.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents[0].Addresses = []string{"erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t", // init sys account
		},
	})
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	nonce := uint64(5)

	bridgeData := DeploySovereignBridgeSetup(t, cs, wallet, esdtSafeWasmPath, feeMarketWasmPath)

	fmt.Println(nodeHandler.GetCoreComponents().AddressPubKeyConverter().SilentEncode(bridgeData.ESDTSafeAddress, log))

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, nodeHandler, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

	logger.SetLogLevel("*:DEBUG")

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: tokenIdentifier,
		Nonce:      0,
		Amount:     amountToDeposit,
	})
	chainSim.Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes)

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[tokenIdentifier].GetValue().String())

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(tokenIdentifier)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)

	cs.GenerateBlocks(20)
}
