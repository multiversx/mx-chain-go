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
)

const (
	simpleEsdtTestWasmPath = "testdata/simple-esdt-test.wasm"
	creatorAddress         = "8049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8"
	burnEndpoint           = "burn"
)

func TestSovereignChainSimulator_CreateNftWithManyQuantityShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	expectedError := "invalid arguments to process built-in function, invalid quantity for esdt type"

	txData := createNftCreateArgs("da2-NFT2-aba42f", uint64(7), big.NewInt(2), core.NonFungibleV2, creatorAddress)
	expectedFullError := fmt.Sprintf("%s %d (%s)", expectedError, core.NonFungibleV2, core.NonFungibleV2.String())
	executeSimpleEsdtOperationWithError(t, txData, expectedFullError)

	expectedFullError = fmt.Sprintf("%s %d (%s)", expectedError, core.DynamicNFT, core.DynamicNFT.String())
	txData = createNftCreateArgs("da3-NFT3-ffa3bd", uint64(7), big.NewInt(2), core.DynamicNFT, creatorAddress)
	executeSimpleEsdtOperationWithError(t, txData, expectedFullError)
}

func TestSovereignChainSimulator_CreateNftWithInvalidTypeShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	expectedError := "invalid arguments to process built-in function, invalid esdt type"

	expectedFullError := fmt.Sprintf("%s %d (%s)", expectedError, core.Fungible, core.Fungible.String())
	txData := createNftCreateArgs("da-TKN-ade731", uint64(0), big.NewInt(1), core.Fungible, creatorAddress)
	executeSimpleEsdtOperationWithError(t, txData, expectedFullError)

	expectedFullError = fmt.Sprintf("%s %d (%s)", expectedError, core.NonFungible, core.NonFungible.String())
	txData = createNftCreateArgs("da-NFT-4f4325", uint64(7), big.NewInt(1), core.NonFungible, creatorAddress)
	executeSimpleEsdtOperationWithError(t, txData, expectedFullError)

	expectedFullError = fmt.Sprintf("%s %d (%s)", expectedError, core.ESDTType(8), core.ESDTType(8).String())
	txData = createNftCreateArgs("da-NFT-beb731", uint64(7), big.NewInt(1), core.ESDTType(8), creatorAddress)
	executeSimpleEsdtOperationWithError(t, txData, expectedFullError)
}

func executeSimpleEsdtOperationWithError(
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	nonce := uint64(0)

	contractAddress := chainSim.DeployContract(t, cs, initialAddrBytes, &nonce, systemContractDeploy, "", simpleEsdtTestWasmPath)

	txResult := chainSim.SendTransaction(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, txData, uint64(20000000))
	chainSim.RequireSignalError(t, txResult, expectedErr)
}

func TestChainSimulator_CreateAndBurnAllEsdtTypes(t *testing.T) {
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress)
	nonce := uint64(0)

	contractAddress := chainSim.DeployContract(t, cs, initialAddrBytes, &nonce, systemContractDeploy, "", simpleEsdtTestWasmPath)

	tokens := createAllEsdtTypes(t, cs, initialAddrBytes, contractAddress, &nonce)
	for _, token := range tokens {
		checkMetaDataInAccounts(t, cs, token, initialAddress, token.Amount)
	}

	burnAllEsdtTypes(t, cs, initialAddrBytes, contractAddress, &nonce, tokens)
	for _, token := range tokens {
		checkMetaDataInAccounts(t, cs, token, initialAddress, big.NewInt(0))
	}
}

func createAllEsdtTypes(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender, contract []byte,
	nonce *uint64,
) []chainSim.ArgsDepositToken {
	tokens := make([]chainSim.ArgsDepositToken, 0)
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab1-TKN-123456",
		Nonce:      uint64(0),
		Amount:     big.NewInt(14556666767),
		Type:       core.Fungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab1-NFTV2-1a2b3c",
		Nonce:      uint64(3),
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab2-DNFT-ead43f",
		Nonce:      uint64(22),
		Amount:     big.NewInt(1),
		Type:       core.DynamicNFT,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab2-SFT-cedd55",
		Nonce:      uint64(345),
		Amount:     big.NewInt(1421),
		Type:       core.SemiFungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab4-DSFT-f6b4c2",
		Nonce:      uint64(88),
		Amount:     big.NewInt(1534),
		Type:       core.DynamicSFT,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab5-META-4b543b",
		Nonce:      uint64(44),
		Amount:     big.NewInt(6231),
		Type:       core.MetaFungible,
	})
	tokens = append(tokens, chainSim.ArgsDepositToken{
		Identifier: "ab5-DMETA-4b543b",
		Nonce:      uint64(721),
		Amount:     big.NewInt(162367),
		Type:       core.DynamicMeta,
	})

	var txData string
	for _, token := range tokens {
		if token.Type == core.Fungible {
			txData = createLocalMintArgs(token.Identifier, token.Amount)
		} else {
			txData = createNftCreateArgs(token.Identifier, token.Nonce, token.Amount, token.Type, creatorAddress)
		}
		chainSim.SendTransactionWithSuccess(t, cs, sender, nonce, contract, chainSim.ZeroValue, txData, uint64(20000000))
	}

	err := cs.GenerateBlocks(3)
	require.Nil(t, err)

	return tokens
}

func createLocalMintArgs(
	token string,
	amount *big.Int,
) string {
	return "local_mint" +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" +
		"@" + hex.EncodeToString(amount.Bytes())
}

func createNftCreateArgs(
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

func burnAllEsdtTypes(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender, contract []byte,
	nonce *uint64,
	tokens []chainSim.ArgsDepositToken,
) {
	for _, token := range tokens {
		if token.Type == core.Fungible {
			chainSim.TransferESDT(t, cs, sender, contract, nonce, token.Identifier, token.Amount, []byte(burnEndpoint))
		} else {
			chainSim.TransferESDTNFT(t, cs, sender, contract, nonce, token.Identifier, token.Nonce, token.Amount, []byte(burnEndpoint))
		}
	}

	err := cs.GenerateBlocks(3)
	require.Nil(t, err)
}

func TestChainSimulator_CreateTokenAndNFTCollectionSameIdentifierAndMakeTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)
	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	bobAddress := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	bobAddrBytes, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(bobAddress)
	chainSim.InitAddressesAndSysAccState(t, cs, initialAddress, bobAddress)
	nonce := uint64(0)
	bobNonce := uint64(0)

	// deploy the whitelisted contract
	contractAddress := chainSim.DeployContract(t, cs, initialAddrBytes, &nonce, systemContractDeploy, "", simpleEsdtTestWasmPath)
	contractAddressEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(contractAddress)
	require.Equal(t, whiteListedAddress, contractAddressEncoded)

	// mint Token
	tokenIdentifier := "s1-TKN-adf412"
	tokenAmount := big.NewInt(125)
	args := createLocalMintArgs(tokenIdentifier, tokenAmount)
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, args, uint64(20000000))
	token := chainSim.ArgsDepositToken{
		Identifier: tokenIdentifier,
		Nonce:      uint64(0),
		Amount:     tokenAmount,
		Type:       core.Fungible,
	}
	checkMetaDataInAccounts(t, cs, token, initialAddress, tokenAmount)

	// create NFT with same collection identifier as Token
	nftCollection := tokenIdentifier
	nftNonce := uint64(5)
	nftIdentifier := nftCollection + "-" + hex.EncodeToString(big.NewInt(int64(nftNonce)).Bytes())
	nftAmount := big.NewInt(1)
	nftType := core.NonFungibleV2
	args = createNftCreateArgs(nftCollection, nftNonce, nftAmount, nftType, hex.EncodeToString(initialAddrBytes))
	chainSim.SendTransactionWithSuccess(t, cs, initialAddrBytes, &nonce, contractAddress, chainSim.ZeroValue, args, uint64(20000000))
	nft := chainSim.ArgsDepositToken{
		Identifier: nftCollection,
		Nonce:      nftNonce,
		Amount:     nftAmount,
		Type:       nftType,
	}
	checkMetaDataInAccounts(t, cs, nft, initialAddress, nftAmount)

	// transfer some Tokens to Bob address
	tokenAmountToSend := big.NewInt(11)
	chainSim.TransferESDT(t, cs, initialAddrBytes, bobAddrBytes, &nonce, tokenIdentifier, tokenAmountToSend)
	receivedToken := chainSim.ArgsDepositToken{
		Identifier: tokenIdentifier,
		Nonce:      uint64(0),
		Amount:     tokenAmountToSend,
		Type:       core.Fungible,
	}
	token.Amount = big.NewInt(0).Sub(token.Amount, receivedToken.Amount)
	checkTokenDataInAccount(t, cs, token, initialAddress, token.Amount)
	checkTokenDataInAccount(t, cs, receivedToken, bobAddress, receivedToken.Amount)

	// transfer NFT to Bob address
	chainSim.TransferESDTNFT(t, cs, initialAddrBytes, bobAddrBytes, &nonce, nftCollection, nftNonce, nftAmount)
	checkTokenDataInAccount(t, cs, nft, initialAddress, big.NewInt(0))
	err = cs.GenerateBlocks(3)
	require.Nil(t, err)
	checkTokenDataInAccount(t, cs, nft, bobAddress, nft.Amount)

	// fast-forward some epochs
	err = cs.GenerateBlocksUntilEpochIsReached(6)
	require.Nil(t, err)

	// transfer NFT back to initial address
	chainSim.TransferESDTNFT(t, cs, bobAddrBytes, initialAddrBytes, &bobNonce, nftCollection, nftNonce, nftAmount)
	checkTokenDataInAccount(t, cs, nft, bobAddress, big.NewInt(0))
	err = cs.GenerateBlocks(3)
	require.Nil(t, err)
	checkTokenDataInAccount(t, cs, nft, initialAddress, nft.Amount)

	// burn some Token
	burnAmount := big.NewInt(3)
	chainSim.TransferESDT(t, cs, initialAddrBytes, contractAddress, &nonce, tokenIdentifier, burnAmount, []byte(burnEndpoint))
	token.Amount = big.NewInt(0).Sub(token.Amount, burnAmount)
	checkTokenDataInAccount(t, cs, token, initialAddress, token.Amount)

	// burn NFT from initial address
	chainSim.TransferESDTNFT(t, cs, initialAddrBytes, contractAddress, &nonce, nftCollection, nftNonce, nftAmount, []byte(burnEndpoint))
	checkTokenDataInAccount(t, cs, nft, initialAddress, big.NewInt(0))

	// the contract should not own any of those tokens
	chainSim.RequireAccountHasToken(t, cs, tokenIdentifier, contractAddressEncoded, big.NewInt(0))
	chainSim.RequireAccountHasToken(t, cs, nftIdentifier, contractAddressEncoded, big.NewInt(0))
}
