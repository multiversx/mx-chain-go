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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
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
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	// Deploy bridge setup on shard 1
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       "sov1",
		IssuePaymentToken: "ABC-123456",
	}
	initOwnerAndSysAccState(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, argsEsdtSafe, enshrineEsdtSafeContract, enshrineEsdtSafeWasmPath)

	addressShardID := chainSim.GetShardForAddress(cs, initialAddress)
	nodeHandler := cs.GetNodeHandler(addressShardID)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	// TODO MX-15942 uncomment after rust framework and contract framework will be updated with all esdt types
	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: argsEsdtSafe.ChainPrefix + "-TKN-123456",
		Nonce:      uint64(0),
		Amount:     big.NewInt(555666),
		Type:       core.Fungible,
	})
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: argsEsdtSafe.ChainPrefix + "-NFTV2-1a2b3c",
		Nonce:      uint64(3),
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})
	//bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
	//	Identifier: argsEsdtSafe.ChainPrefix + "-DNFT-ead43f",
	//	Nonce:      uint64(6),
	//	Amount:     big.NewInt(1),
	//	Type:       core.DynamicNFT,
	//})
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: argsEsdtSafe.ChainPrefix + "-SFT-cedd55",
		Nonce:      uint64(52),
		Amount:     big.NewInt(2156),
		Type:       core.SemiFungible,
	})
	//bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
	//	Identifier: argsEsdtSafe.ChainPrefix + "-DSFT-f6b4c2",
	//	Nonce:      uint64(33),
	//	Amount:     big.NewInt(11),
	//	Type:       core.DynamicSFT,
	//})
	//bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
	//	Identifier: argsEsdtSafe.ChainPrefix + "-META-4b543b",
	//	Nonce:      uint64(57),
	//	Amount:     big.NewInt(124000),
	//	Type:       core.MetaFungible,
	//})
	//bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
	//	Identifier: argsEsdtSafe.ChainPrefix + "-DMETA-4b543b",
	//	Nonce:      uint64(31),
	//	Amount:     big.NewInt(14326743),
	//	Type:       core.DynamicMeta,
	//})

	// Register the sovereign tokens in contract
	tokens := getUniquePrefixedTokens(bridgedInTokens, argsEsdtSafe.ChainPrefix)
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, tokens)

	// We will deposit prefixed tokens from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our address from another shard.
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, wallet.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	for _, bridgedInToken := range groupTokens(bridgedInTokens) {
		checkTokenDataInAccount(t, cs, bridgedInToken, esdtSafeEncoded, big.NewInt(0))       // sender
		checkTokenDataInAccount(t, cs, bridgedInToken, wallet.Bech32, bridgedInToken.Amount) // receiver
	}

	// Generate address in another shard than bridge contract and user wallet
	receiver, err := cs.GenerateAndMintWalletAddress(2, chainSim.InitialAmount)
	require.Nil(t, err)
	receiverNonce := uint64(0)
	receiver2, err := cs.GenerateAndMintWalletAddress(2, chainSim.InitialAmount)
	require.Nil(t, err)

	// will make a cross shard transaction, then deposit transaction, then same shard transaction with every token
	for _, token := range groupTokens(bridgedInTokens) {
		amountToSend := big.NewInt(2)
		if token.Type == core.NonFungibleV2 || token.Type == core.DynamicNFT || token.Amount.Cmp(big.NewInt(1)) == 0 {
			amountToSend = big.NewInt(1)
		}
		// cross shard transfer
		receiverToken := transferToken(t, cs, token, wallet, receiver, &nonce, amountToSend)
		checkTokenDataInAccount(t, cs, token, wallet.Bech32, big.NewInt(0).Sub(token.Amount, receiverToken.Amount))
		checkTokenDataInAccount(t, cs, token, receiver.Bech32, receiverToken.Amount)

		amountToDeposit := big.NewInt(1)
		bridgedOutToken := chainSim.ArgsDepositToken{
			Identifier: token.Identifier,
			Nonce:      token.Nonce,
			Amount:     amountToDeposit,
			Type:       token.Type,
		}

		// deposit a prefixed token from main chain to sovereign chain,
		// expecting these tokens to be burned by the whitelisted ESDT safe sc
		txResult = deposit(t, cs, receiver.Bytes, &receiverNonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{bridgedOutToken}, receiver.Bytes)
		chainSim.RequireSuccessfulTransaction(t, txResult)

		remainingTokenAmount := big.NewInt(0).Sub(receiverToken.Amount, amountToDeposit)
		checkTokenDataInAccount(t, cs, token, receiver.Bech32, remainingTokenAmount)
		checkTokenDataInAccount(t, cs, token, esdtSafeEncoded, big.NewInt(0))
		receiverToken.Amount = remainingTokenAmount

		if remainingTokenAmount.Cmp(big.NewInt(0)) > 0 {
			// same shard transfer
			transferredToken := transferToken(t, cs, receiverToken, receiver, receiver2, &receiverNonce, remainingTokenAmount)
			checkTokenDataInAccount(t, cs, transferredToken, receiver.Bech32, big.NewInt(0))
			checkTokenDataInAccount(t, cs, transferredToken, receiver2.Bech32, remainingTokenAmount)
		}
	}
}

func transferToken(
	t *testing.T,
	cs chainSim.ChainSimulator,
	token chainSim.ArgsDepositToken,
	sender, receiver dtos.WalletAddress,
	senderNonce *uint64,
	amountToSend *big.Int,
) chainSim.ArgsDepositToken {
	if token.Type == core.Fungible {
		chainSim.TransferESDT(t, cs, sender.Bytes, receiver.Bytes, senderNonce, token.Identifier, amountToSend)
	} else {
		chainSim.TransferESDTNFT(t, cs, sender.Bytes, receiver.Bytes, senderNonce, token.Identifier, token.Nonce, amountToSend)
		err := cs.GenerateBlocks(3)
		require.Nil(t, err)
	}

	return chainSim.ArgsDepositToken{
		Identifier: token.Identifier,
		Nonce:      token.Nonce,
		Amount:     amountToSend,
		Type:       token.Type,
	}
}

func checkTokenDataInAccount(
	t *testing.T,
	cs chainSim.ChainSimulator,
	token chainSim.ArgsDepositToken,
	address string,
	expectedAmount *big.Int,
) {
	chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), address, expectedAmount)
	checkMetaDataInAccounts(t, cs, token, address, expectedAmount)
}
