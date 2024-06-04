package bridge

import (
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
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
	esdtSafeWasmPath           = "testdata/esdt-safe.wasm"
	feeMarketWasmPath          = "testdata/fee-market.wasm"
)

// TODO: MX-15527 Make a similar bridge test with sovereign chain simulator after merging this into feat/chain-go-sdk

// In this test we:
// - will deploy sovereign bridge contracts on the main chain
// - will whitelist the bridge esdt safe contract to allow it to burn/mint cross chain esdt tokens
// - deposit a cross chain prefixed token in the contract to be transferred to an address (mint)
// - send back some of the cross chain prefixed tokens to the contract to be bridged to a sovereign chain (burn)
// - move some of the prefixed tokens to another address
func TestChainSimulator_ExecuteMintBurnBridgeOpForESDTTokensWithPrefix(t *testing.T) {
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
	depositToken := "sov1-SOVT-5d8f56"
	expectedMintValue, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		"00000001" + // nr of tokens
		"00000010" + // length of token identifier
		hex.EncodeToString([]byte(depositToken)) + //token identifier
		"0000000000000000000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" + // nonce + token data
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(50000000))
	chainSim.RequireAccountHasToken(t, cs, depositToken, wallet.Bech32, expectedMintValue)

	amountToDeposit, _ := big.NewInt(0).SetString("120000000000000000000", 10)
	remainingValueAfterBridge := expectedMintValue.Sub(expectedMintValue, amountToDeposit)
	depositTokens := make([]chainSim.ArgsDepositToken, 0)
	depositTokens = append(depositTokens, chainSim.ArgsDepositToken{
		Identifier: depositToken,
		Nonce:      0,
		Amount:     amountToDeposit,
	})
	Deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, depositTokens, wallet.Bytes)
	chainSim.RequireAccountHasToken(t, cs, depositToken, wallet.Bech32, remainingValueAfterBridge)

	// Send some of the bridged prefixed tokens to another address
	receiver := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	receiverBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(receiver)
	require.Nil(t, err)

	receivedTokens := big.NewInt(11)
	chainSim.TransferESDT(t, cs, wallet.Bytes, receiverBytes, &nonce, depositToken, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, depositToken, receiver, receivedTokens)
	chainSim.RequireAccountHasToken(t, cs, depositToken, initialAddress, big.NewInt(0).Sub(remainingValueAfterBridge, receivedTokens))
}
