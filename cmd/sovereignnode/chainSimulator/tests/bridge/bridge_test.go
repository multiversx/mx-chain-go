package bridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
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
	multiSigWasmPath           = "../testdata/multisigverifier.wasm"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))
var issueCost = big.NewInt(5000000000000000000)

func TestBridge_DeployOnMainChain_IssueAndDeposit(t *testing.T) {
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  uint64(6000),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		MetaChainConsensusGroupSize: 1,
		ConsensusGroupSize:          1,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	systemEsdtScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")
	nonce := uint64(0)

	wallet, err := cs.GenerateAndMintWalletAddress(0, initialMinting)
	require.Nil(t, err)

	esdtSafeArgs := "@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setFeeMarketAddressData, uint64(10000000))

	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))

	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	multiSigAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", multiSigWasmPath)

	setEsdtSafeAddressData := "setEsdtSafeAddress" +
		"@" + hex.EncodeToString(esdtSafeAddress)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, multiSigAddress, big.NewInt(0), setEsdtSafeAddressData, uint64(10000000))

	setMultiSigAddressData := "setMultisigAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setMultiSigAddressData, uint64(10000000))

	setSovereignBridgeAddressData := "setSovereignBridgeAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setSovereignBridgeAddressData, uint64(10000000))

	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte("MainToken")) +
		"@" + hex.EncodeToString([]byte("TKN")) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", 18) // num decimals
	issueTx := chainSimulatorIntegrationTests.SendTransactionAndWaitForMaxNrOfBlocks(t, cs, wallet.Bytes, &nonce, systemEsdtScAddress, issueCost, issueArgs, uint64(60000000), 3) // 3 blocks to execute on destination shard (meta)
	tokenIdentifier := issueTx.Logs.Events[0].Topics[0]
	require.True(t, len(tokenIdentifier) > 7)

	err = cs.GenerateBlocks(1) // one more block to get the tokens in source shard
	require.Nil(t, err)

	tokens, _, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 1)
	require.Equal(t, supply.String(), tokens[string(tokenIdentifier)].GetValue().String())

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositArgs := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(esdtSafeAddress) +
		"@01" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + // nonce
		"@" + hex.EncodeToString(amountToDeposit.Bytes()) +
		"@" + hex.EncodeToString([]byte("deposit")) + // deposit func
		"@" + hex.EncodeToString(wallet.Bytes) //receiver from other side
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))

	tokens, _, err = cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 1)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[string(tokenIdentifier)].GetValue().String())
}

func TestBridge_DeployOnSovereignChain_IssueAndDeposit(t *testing.T) {
	scs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ChainSimulatorArgs: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, scs)

	defer scs.Close()

	err = scs.GenerateBlocks(200)
	require.Nil(t, err)

	sovNodeHandler := scs.GetNodeHandler(core.SovereignChainShardId)
	sovSystemScAddress, _ := sovNodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	sovSystemEsdtScAddress, _ := sovNodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")
	sovNonce := uint64(0)

	sovWallet, err := scs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeSovArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(sovWallet.Bytes) // initiator_address
	esdtSafeSovAddress := chainSimulatorIntegrationTests.DeployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, esdtSafeSovArgs, esdtSafeWasmPath)

	feeMarketSovArgs := "@" + hex.EncodeToString(esdtSafeSovAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketSovAddress := chainSimulatorIntegrationTests.DeployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, feeMarketSovArgs, feeMarketWasmPath)

	setSovFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketSovAddress)
	chainSimulatorIntegrationTests.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, big.NewInt(0), setSovFeeMarketAddressData, uint64(10000000))

	chainSimulatorIntegrationTests.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, feeMarketSovAddress, big.NewInt(0), "disableFee", uint64(10000000))

	chainSimulatorIntegrationTests.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, big.NewInt(0), "unpause", uint64(10000000))

	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte("SovToken")) +
		"@" + hex.EncodeToString([]byte("SVN")) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", 18) // num decimals
	txResult := chainSimulatorIntegrationTests.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, sovSystemEsdtScAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]
	require.True(t, len(tokenIdentifier) > 7)

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositArgs := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(esdtSafeSovAddress) +
		"@01" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + // nonce
		"@" + hex.EncodeToString(amountToDeposit.Bytes()) +
		"@" + hex.EncodeToString([]byte("deposit")) + // deposit func
		"@" + hex.EncodeToString(sovWallet.Bytes) //receiver from other side
	chainSimulatorIntegrationTests.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, sovWallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))

	tokens, _, err := sovNodeHandler.GetFacadeHandler().GetAllESDTTokens(sovWallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[string(tokenIdentifier)].GetValue().String())
}
