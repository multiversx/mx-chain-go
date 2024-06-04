package bridge

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

const (
	issuePrice = "5000000000000000000"
)

func TestChainSimulator_IssueBurnNft(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(0)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
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

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	nonce := uint64(0)

	bridgeData := DeployBridgeSetup(t, cs, wallet.Bytes, &nonce, esdtSafeWasmPath, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, esdtSafeEncoded, whiteListedAddress)

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our address.
	depositToken := "sov1-SOVNFT-123456"
	expectedMintValue := big.NewInt(1)
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		"00000001" + // nr of tokens
		lengthOn4Bytes(len(depositToken)) + // length of token identifier
		hex.EncodeToString([]byte(depositToken)) + //token identifier
		"0000000000000001" + // nonce
		"01" + // NFT type
		lengthOn4Bytes(len(expectedMintValue.Bytes())) + // length of amount
		hex.EncodeToString(expectedMintValue.Bytes()) + // amount
		"00" + // frozen
		"00000000" + // length of hash
		"00000003" + // length of name
		hex.EncodeToString([]byte("ABC")) + // name
		"00000000" + // length of attributes
		"0000000000000000000000000000000000000000000000000000000000000000" + // creator
		"00000002" + // length of royalties
		hex.EncodeToString(big.NewInt(1000).Bytes()) +
		"00000000" + // length of uris
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(50000000))
	chainSim.RequireAccountHasToken(t, cs, depositToken+"-01", wallet.Bech32, expectedMintValue)

	amountToDeposit := big.NewInt(1)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: depositToken,
		Nonce:      1,
		Amount:     amountToDeposit,
	})
	Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes)

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Empty(t, esdts)

	esdts, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(esdtSafeEncoded, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Empty(t, esdts)

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(depositToken)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, "1", tokenSupply.Burned)
}

func TestChainSimulator_IssueBurnSft(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = issuePrice
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(0)

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
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

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	nonce := uint64(0)

	bridgeData := DeployBridgeSetup(t, cs, wallet.Bytes, &nonce, esdtSafeWasmPath, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, esdtSafeEncoded, whiteListedAddress)

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our address.
	depositToken := "sov1-SOVSFT-654321"
	expectedMintValue := big.NewInt(50)
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		"00000001" + // nr of tokens
		lengthOn4Bytes(len(depositToken)) + // length of token identifier
		hex.EncodeToString([]byte(depositToken)) + //token identifier
		"0000000000000001" + // nonce
		"03" + // SFT type
		lengthOn4Bytes(len(expectedMintValue.Bytes())) + // length of amount
		hex.EncodeToString(expectedMintValue.Bytes()) + // amount
		"00" + // frozen
		"00000000" + // length of hash
		"00000003" + // length of name
		hex.EncodeToString([]byte("ABC")) + // name
		"00000000" + // length of attributes
		"0000000000000000000000000000000000000000000000000000000000000000" + // creator
		"00000002" + // length of royalties
		hex.EncodeToString(big.NewInt(1000).Bytes()) +
		"00000000" + // length of uris
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(50000000))
	chainSim.RequireAccountHasToken(t, cs, depositToken+"-01", wallet.Bech32, expectedMintValue)

	amountToBurn := big.NewInt(10)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: depositToken,
		Nonce:      1,
		Amount:     amountToBurn,
	})
	Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes)

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.Equal(t, 1, len(esdts))
	require.Equal(t, big.NewInt(0).Sub(expectedMintValue, amountToBurn).String(), esdts[depositToken+"-01"].Value)

	esdts, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(esdtSafeEncoded, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Empty(t, esdts)

	tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(depositToken)
	require.Nil(t, err)
	require.NotNil(t, tokenSupply)
	require.Equal(t, amountToBurn.String(), tokenSupply.Burned)
}

func lengthOn4Bytes(number int) string {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(number))
	return hex.EncodeToString(bytes)
}
