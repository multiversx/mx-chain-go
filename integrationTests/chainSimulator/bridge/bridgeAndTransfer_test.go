package bridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

// TODO: MX-15527 Make a similar bridge test with sovereign chain simulator after merging this into feat/chain-go-sdk

// In this test we:
// - will deploy sovereign bridge contracts on the main chain
// - will whitelist the bridge esdt safe contract to allow it to burn/mint cross chain esdt tokens
// - deposit a cross chain prefixed token in the contract to be transferred to an address (mint)
// - send back some of the cross chain prefixed tokens to the contract to be bridged to a sovereign chain (burn)
// - move some of the prefixed tokens to another address
func TestChainSimulator_ExecuteMintBurnBridgeOpForESDTTokensWithPrefixAndTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              1,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    uint64(6000),
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)

	// Deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       "sov1",
		IssuePaymentToken: "ABC-123456",
	}
	initOwnerAndSysAccState(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, esdtSafeWasmPath, argsEsdtSafe, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	tokenToBridge := argsEsdtSafe.ChainPrefix + "-SOVT-5d8f56"
	expectedMintValue := big.NewInt(123)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: tokenToBridge,
		Nonce:      0,
		Amount:     expectedMintValue,
		Type:       core.Fungible,
	})

	// Register the sovereign token in contract
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, []string{tokenToBridge})

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our address.
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, wallet.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	for _, token := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), wallet.Bech32, token.Amount)
	}

	amountToDeposit := big.NewInt(23)
	bridgedOutTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: tokenToBridge,
		Nonce:      0,
		Amount:     amountToDeposit,
		Type:       core.Fungible,
	})
	// We will deposit a prefixed token from main chain to sovereign chain,
	// expecting these tokens to be burned by the whitelisted ESDT safe sc
	txResult = Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, bridgedOutTokens, wallet.Bytes)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	remainingValueAfterBridge := big.NewInt(0).Sub(expectedMintValue, amountToDeposit)
	chainSim.RequireAccountHasToken(t, cs, tokenToBridge, wallet.Bech32, remainingValueAfterBridge)
	chainSim.RequireAccountHasToken(t, cs, tokenToBridge, esdtSafeEncoded, big.NewInt(0))

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(tokenToBridge)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToDeposit.String(), tokenSupply.Burned)

	// Send some of the bridged prefixed tokens to another address
	receiver := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	receiverBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(receiver)
	require.Nil(t, err)

	receivedTokens := big.NewInt(11)
	chainSim.TransferESDT(t, cs, wallet.Bytes, receiverBytes, &nonce, tokenToBridge, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, tokenToBridge, receiver, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, tokenToBridge, wallet.Bech32, big.NewInt(0).Sub(remainingValueAfterBridge, receivedTokens))
}
