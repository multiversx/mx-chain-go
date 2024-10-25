package bridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

// 1. transfer from sovereign chain to main chain
// 2. transfer from main chain to sovereign chain
// 3. transfer again the same tokens from sovereign chain to main chain
// tokens are originated from sovereign chain
// esdt-safe contract in main chain will issue its own tokens at registerToken step
// and will work with a mapper between sov_id <-> main_id
func TestChainSimulator_ExecuteWithTransferDataFails(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    uint64(6000),
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePaymentCost
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	// deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, ArgsEsdtSafe{}, esdtSafeContract, esdtSafeWasmPath)
	esdtSafeAddr, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	esdtSafeAddrShard := chainSim.GetShardForAddress(cs, esdtSafeAddr)

	nodeHandler := cs.GetNodeHandler(esdtSafeAddrShard)

	wallet, err := cs.GenerateAndMintWalletAddress(esdtSafeAddrShard, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)
	helloAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemContractDeploy, "", helloWasmPath)
	helloAddrBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(helloAddress)

	tokens := make([]chainSim.ArgsDepositToken, 0)
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab1-TKN-fasd35",
		Nonce:      uint64(0),
		Amount:     big.NewInt(14556666767),
		Type:       core.Fungible,
	})

	tokensMapper := make(map[string]string)

	// transfer sovereign chain -> main chain with transfer data
	// token originated from sovereign chain
	for _, token := range tokens {
		// register sovereign token identifier
		// this will issue a new token on main chain and create a mapper between the identifiers
		registerTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, token)
		tokensMapper[token.Identifier] = chainSim.GetIssuedEsdtIdentifier(t, cs.GetNodeHandler(core.MetachainShardId), getTokenTicker(token.Identifier), token.Type.String())

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred with SC call to hello contract
		trnsData := &transferData{
			GasLimit: uint64(10000000),
			Function: []byte("hello"),
			Args:     [][]byte{{0x00}},
		}
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, helloAddress, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, trnsData)
		// expecting the operation to fail in hello contract with error "Value should be greater than 0"
		// but receive the callback in esdt-safe contract
		chainSim.RequireSuccessfulTransaction(t, txResult)
		receivedToken := chainSim.ArgsDepositToken{
			Identifier: tokensMapper[token.Identifier],
			Nonce:      token.Nonce,
			Amount:     token.Amount,
			Type:       token.Type,
		}

		// no tokens in contracts
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(0))
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), helloAddrBech32, big.NewInt(0))

		// tokens should be burned because operation failed
		_ = cs.GenerateBlocks(5)
		tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(getTokenIdentifier(receivedToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, receivedToken.Amount.String(), tokenSupply.Burned)
	}
}
