package bridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/vm"

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
	issuePrice                 = "5000000000000000000"
)

func TestBridge_DeployOnMainChain_IssueAndDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)

	systemScAddress, err := chainSim.GetSysAccBytesAddress(nodeHandler)
	require.Nil(t, err)

	nonce := uint64(0)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)

	esdtSafeArgs := "@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	multiSigAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", multiSigWasmPath)

	setEsdtSafeAddressData := "setEsdtSafeAddress" +
		"@" + hex.EncodeToString(esdtSafeAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, multiSigAddress, big.NewInt(0), setEsdtSafeAddressData, uint64(10000000))

	setMultiSigAddressData := "setMultisigAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setMultiSigAddressData, uint64(10000000))

	setSovereignBridgeAddressData := "setSovereignBridgeAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setSovereignBridgeAddressData, uint64(10000000))

	issueCost, _ := big.NewInt(0).SetString(issuePrice, 10)
	supply, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	tokenName := "MainToken"
	tokenTicker := "TKN"
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte(tokenName)) +
		"@" + hex.EncodeToString([]byte(tokenTicker)) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", 18) // num decimals
	issueTx := chainSim.GenerateTransaction(wallet.Bytes, nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(issueTx, 3) // 3 blocks to execute on destination shard (meta)
	nonce++
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]
	require.Equal(t, len(tokenTicker)+7, len(tokenIdentifier))
	require.Equal(t, tokenName, string(txResult.Logs.Events[2].Topics[1]))
	require.Equal(t, tokenTicker, string(txResult.Logs.Events[2].Topics[2]))

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
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))

	tokens, _, err = cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 1)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[string(tokenIdentifier)].GetValue().String())
}

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
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte("SovToken")) +
		"@" + hex.EncodeToString([]byte("SVN")) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", 18) // num decimals
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]
	require.True(t, len(tokenIdentifier) > 7)

	amountToDeposit, _ := big.NewInt(0).SetString("2000000000000000000", 10)
	depositArgs := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(bridgeData.ESDTSafeAddress) +
		"@01" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + // nonce
		"@" + hex.EncodeToString(amountToDeposit.Bytes()) +
		"@" + hex.EncodeToString([]byte("deposit")) + // deposit func
		"@" + hex.EncodeToString(wallet.Bytes) //receiver from other side
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
	require.Equal(t, big.NewInt(0).Sub(supply, amountToDeposit).String(), tokens[string(tokenIdentifier)].GetValue().String())
}
