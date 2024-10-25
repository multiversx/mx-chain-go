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

// transfer from sovereign chain to main chain with transfer data
// tokens are originated from sovereign chain
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

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	receiverShardId := uint32(0)

	// TODO MX-15942 uncomment dynamic tokens, currently there is no issue function in SC framework for dynamic esdts
	tokens := make([]chainSim.ArgsDepositToken, 0)
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab1-TKN-fasd35",
		Nonce:      uint64(0),
		Amount:     big.NewInt(14556666767),
		Type:       core.Fungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab1-NFTV2-1ds234",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})
	//tokens = append(tokens, chainSim.ArgsDepositToken{
	//	Identifier: "ab2-DNFT-fdfe3r",
	//	Nonce:      uint64(1),
	//	Amount:     big.NewInt(1),
	//	Type:       core.DynamicNFT,
	//})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab2-SFT-gw4fw2",
		Nonce:      uint64(1),
		Amount:     big.NewInt(1421),
		Type:       core.SemiFungible,
	})
	//tokens = append(tokens, chainSim.ArgsDepositToken{
	//	Identifier: "ab4-DSFT-g43g2s",
	//	Nonce:      uint64(1),
	//	Amount:     big.NewInt(1534),
	//	Type:       core.DynamicSFT,
	//})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab5-META-1ds234",
		Nonce:      uint64(1),
		Amount:     big.NewInt(6231),
		Type:       core.MetaFungible,
	})
	//tokens = append(tokens, chainSim.ArgsDepositToken{
	//	Identifier: "ab5-DMETA-f23g2f",
	//	Nonce:      uint64(1),
	//	Amount:     big.NewInt(162367),
	//	Type:       core.DynamicMeta,
	//})

	// generate hello contracts in each shard
	// hello contract has one endpoint "hello" which receives a number as argument
	// if number is 0 then will throw error, otherwise will do nothing
	receiverContracts := deployReceiverContractInAllShards(t, cs)

	tokensMapper := make(map[string]string)

	// transfer sovereign chain -> main chain -> sovereign chain
	// token originated from sovereign chain
	for _, token := range tokens {
		// register sovereign token identifier
		// this will issue a new token on main chain and create a mapper between the identifiers
		registerTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, token)
		tokensMapper[token.Identifier] = chainSim.GetIssuedEsdtIdentifier(t, cs.GetNodeHandler(core.MetachainShardId), getTokenTicker(token.Identifier), token.Type.String())

		// get contract from next shard
		receiver := receiverContracts[receiverShardId]

		// execute operations received from sovereign chain
		// expecting the token to be minted in esdt-safe contract with the same properties and transferred with SC call to hello contract
		// -------------
		// for (dynamic) SFT/MetaESDT the contract will create one more token and keep it forever
		// because if the same token is received 2nd time, the contract will just add quantity, not create different token
		trnsData := &transferData{
			GasLimit: uint64(10000000),
			Function: []byte("hello"),
			Args:     [][]byte{{0x00}},
		}
		// the executed operation in hello contract is expected to fail, tokens will be minted and then burned
		txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, receiver.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, []chainSim.ArgsDepositToken{token}, wallet.Bytes, trnsData)
		chainSim.RequireSuccessfulTransaction(t, txResult)
		receivedToken := chainSim.ArgsDepositToken{
			Identifier: tokensMapper[token.Identifier],
			Nonce:      token.Nonce,
			Amount:     token.Amount,
			Type:       token.Type,
		}
		if isSftOrMeta(receivedToken.Type) {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(1))
		} else {
			chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), esdtSafeAddr, big.NewInt(0))
		}

		waitIfCrossShardProcessing(cs, esdtSafeAddrShard, receiverShardId)
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(receivedToken), receiver.Bech32, big.NewInt(0))

		_ = cs.GenerateBlocks(5)
		tokenSupply, err := cs.GetNodeHandler(esdtSafeAddrShard).GetFacadeHandler().GetTokenSupply(getTokenIdentifier(receivedToken))
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, receivedToken.Amount.String(), tokenSupply.Burned)

		nextShardId(&receiverShardId)
	}
}
