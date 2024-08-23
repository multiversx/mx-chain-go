package bridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	simpleEsdtTestWasmPath = "testdata/simple-esdt-test.wasm"
	creatorAddress         = "8049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8"
)

func TestSovereignChainSimulator_CreateNftWithManyQuantityShouldFail(t *testing.T) {
	expectedError := "invalid arguments to process built-in function, invalid quantity"
	txData := createNftArgs("da1-NFT1-waf332", uint64(7), big.NewInt(2), core.NonFungible, creatorAddress)
	executeSimpleEsdtOperation(t, txData, expectedError)

	txData = createNftArgs("da2-NFT2-geg42g", uint64(7), big.NewInt(2), core.NonFungibleV2, creatorAddress)
	executeSimpleEsdtOperation(t, txData, expectedError)

	txData = createNftArgs("da3-NFT3-gew3gr", uint64(7), big.NewInt(2), core.DynamicNFT, creatorAddress)
	executeSimpleEsdtOperation(t, txData, expectedError)
}

func TestSovereignChainSimulator_CreateNftWithInvalidTypeShouldFail(t *testing.T) {
	expectedError := "invalid arguments to process built-in function, invalid esdt type"
	txData := createNftArgs("da-NFT-ten731", uint64(7), big.NewInt(1), core.ESDTType(8), creatorAddress)
	executeSimpleEsdtOperation(t, txData, expectedError)
}

func executeSimpleEsdtOperation(
	t *testing.T,
	txData string,
	expectedErr string,
) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: esdtSystemAccount, // init sys account
		},
	})
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	contractAddress := chainSim.DeployContract(t, cs, initialAddrBytes, &nonce, systemContractDeploy, "", simpleEsdtTestWasmPath)

	txResult := chainSim.SendTransaction(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, txData, uint64(20000000))
	chainSim.RequireSignalError(t, txResult, expectedErr)
}

func TestChainSimulator_CreateTokenAndNFTCollectionSameIdentifier(t *testing.T) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: esdtSystemAccount, // init sys account
		},
	})
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// deploy the whitelisted contract
	contractAddress := chainSim.DeployContract(t, cs, initialAddrBytes, &nonce, systemContractDeploy, "", simpleEsdtTestWasmPath)
	contractAddressEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(contractAddress)
	require.Equal(t, whiteListedAddress, contractAddressEncoded)

	// mint Token
	tokenIdentifier := "s1-TKN-gdf412"
	tokenAmount := big.NewInt(125)
	args := createLocalMintArgs(tokenIdentifier, uint64(0), tokenAmount)
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, args, uint64(20000000))
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, initialAddress, tokenAmount)

	// create NFT with same collection identifier as Token
	nftCollection := tokenIdentifier
	nftNonce := uint64(5)
	nftIdentifier := createEsdtIdentifier(nftCollection, nftNonce)
	nftAmount := big.NewInt(1)
	args = createNftArgs(nftCollection, nftNonce, nftAmount, core.NonFungibleV2, hex.EncodeToString(initialAddrBytes))
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, args, uint64(20000000))
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, initialAddress, nftAmount)

	bobAddress := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	bobAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(bobAddress)
	require.Nil(t, err)
	bobNonce := uint64(0)
	// get some funds in the bobAddress wallet
	chainSim.SendTransaction(t, cs, initialAddrBytes, &nonce, bobAddrBytes, chainSim.OneEGLD, "", uint64(50000))

	// transfer some Tokens to Bob address
	receivedTokens := big.NewInt(11)
	chainSim.TransferESDT(t, cs, initialAddrBytes, bobAddrBytes, &nonce, tokenIdentifier, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, bobAddress, receivedTokens)
	tokenAmount = big.NewInt(0).Sub(tokenAmount, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, initialAddress, tokenAmount)

	// transfer NFT to Bob address
	chainSim.TransferESDTNFT(t, cs, initialAddrBytes, bobAddrBytes, &nonce, nftCollection, nftNonce, nftAmount)
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, bobAddress, nftAmount)
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, initialAddress, big.NewInt(0))

	// transfer NFT back to initial address
	chainSim.TransferESDTNFT(t, cs, bobAddrBytes, initialAddrBytes, &bobNonce, nftCollection, nftNonce, nftAmount)
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, initialAddress, nftAmount)
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, bobAddress, big.NewInt(0))

	// burn some Token
	burnAmount := big.NewInt(3)
	args = createBurnTokenArgs(tokenIdentifier, burnAmount)
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, args, uint64(20000000))
	tokenAmount = big.NewInt(0).Sub(tokenAmount, burnAmount)
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, initialAddress, tokenAmount)

	// burn NFT
	args = createBurnNftArgs(nftCollection, nftNonce, nftAmount, contractAddress)
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, initialAddrBytes, chainSim.ZeroValue, args, uint64(20000000))
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, initialAddress, big.NewInt(0))

	// the contract should not own any of those tokens
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, contractAddressEncoded, big.NewInt(0))
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, contractAddressEncoded, big.NewInt(0))
}

func createLocalMintArgs(
	token string,
	nonce uint64,
	amount *big.Int,
) string {
	return "local_mint" +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" + getTokenNonce(nonce) +
		"@" + hex.EncodeToString(amount.Bytes())
}

func createNftArgs(
	identifier string,
	nonce uint64,
	amount *big.Int,
	esdtType core.ESDTType,
	creator string,
) string {
	return "nft_create" +
		"@" + hex.EncodeToString([]byte(identifier)) +
		"@" + getTokenNonce(nonce) +
		"@" + hex.EncodeToString(amount.Bytes()) +
		"@" + fmt.Sprintf("%02x", uint32(esdtType)) +
		"@" + creator
}

func createBurnTokenArgs(
	token string,
	amount *big.Int,
) string {
	return core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" + hex.EncodeToString(amount.Bytes()) +
		"@" + hex.EncodeToString([]byte("burn"))
}

func createBurnNftArgs(
	token string,
	nonce uint64,
	amount *big.Int,
	contract []byte,
) string {
	return core.BuiltInFunctionESDTNFTTransfer +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" + getTokenNonce(nonce) +
		"@" + hex.EncodeToString(amount.Bytes()) +
		"@" + hex.EncodeToString(contract) +
		"@" + hex.EncodeToString([]byte("burn"))
}
