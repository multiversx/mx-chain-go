package bridge

import (
	"math/big"
	"testing"
	"time"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
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

// This test will:
// - deploy bridge contracts setup
// - issue a new fungible token
// - deposit some tokens in esdt-safe contract
// - check the sender balance is correct
// - check the token supply is correct after deposit (burned amount, supply etc)
func TestBridge_DeployOnSovereignChain_IssueAndDeposit(t *testing.T) {
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
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	time.Sleep(time.Second) // wait for VM to be ready for processing queries

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	bridgeData := DeploySovereignBridgeSetup(t, cs, esdtSafeWasmPath, feeMarketWasmPath)

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "SovToken"
	tokenTicker := "SVN"
	numDecimals := 18
	tokenIdentifier := chainSim.IssueFungible(t, cs, wallet.Bytes, &nonce, issueCost, tokenName, tokenTicker, numDecimals, supply)

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
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[string(tokenIdentifier)].GetValue().String())

	_ = cs.GenerateBlocks(10)

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(string(tokenIdentifier))
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, supply.String(), tokenSupply.InitialMinted)
	require.Equal(t, big.NewInt(0), tokenSupply.Minted)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokenSupply.Supply)
}
