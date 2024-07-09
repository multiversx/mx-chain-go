package bridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
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

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
	esdtSafeWasmPath           = "testdata/esdt-safe.wasm"
	feeMarketWasmPath          = "testdata/fee-market.wasm"
)

var sovChainPrefix = "sov"

func TestChainSimulator_ExecuteWithMintAndBurnFungibleWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	token := sovChainPrefix + "-SOVTKN-1a2b3c"
	tokenNonce := uint64(0)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(123),
		Type:       core.Fungible,
	})
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(100),
		Type:       core.Fungible,
	})

	bridgedOutTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(12),
		Type:       core.Fungible,
	})
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      tokenNonce,
		Amount:     big.NewInt(10),
		Type:       core.Fungible,
	})

	simulateExecutionAndDeposit(t, bridgedInTokens, bridgedOutTokens)
}

func TestChainSimulator_ExecuteWithMintMultipleEsdtsAndBurnNftWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nft := sovChainPrefix + "-SOVNFT-123456"
	nftNonce := uint64(3)
	token := sovChainPrefix + "-TKN-1q2w3e"

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: nft,
		Nonce:      nftNonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungible,
	})
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      0,
		Amount:     big.NewInt(1),
		Type:       core.Fungible,
	})

	bridgedOutTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: nft,
		Nonce:      nftNonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungible,
	})

	simulateExecutionAndDeposit(t, bridgedInTokens, bridgedOutTokens)
}

func TestChainSimulator_ExecuteWithMintAndBurnSftWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sft := sovChainPrefix + "-SOVSFT-654321"
	sftNonce := uint64(123)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: sft,
		Nonce:      sftNonce,
		Amount:     big.NewInt(50),
		Type:       core.SemiFungible,
	})

	bridgedOutTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: sft,
		Nonce:      sftNonce,
		Amount:     big.NewInt(20),
		Type:       core.SemiFungible,
	})

	simulateExecutionAndDeposit(t, bridgedInTokens, bridgedOutTokens)
}

func simulateExecutionAndDeposit(
	t *testing.T,
	bridgedInTokens []chainSim.ArgsDepositToken,
	bridgedOutTokens []chainSim.ArgsDepositToken,
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
		ChainPrefix:       sovChainPrefix,
		IssuePaymentToken: "WEGLD-bd4d79",
	}
	bridgeData := deployBridgeSetup(t, cs, initialAddress, esdtSafeWasmPath, argsEsdtSafe, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	getIssuePaymentTokenInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, paymentTokenAmount)

	// We need to register tokens originated from sovereign (to pay the issue cost)
	// Only the tokens with sovereign prefix need to be registered (these are the ones that will be minted), the rest will be taken from contract balance
	tokens := getSovereignTokens(bridgedInTokens, argsEsdtSafe.ChainPrefix)
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, tokens)

	// We will deposit an array of prefixed tokens from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our wallet address.
	executeMintOperation(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, bridgedInTokens)

	// deposit an array of tokens from main chain to sovereign chain,
	// expecting these tokens to be burned by the whitelisted ESDT safe sc
	deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, bridgedOutTokens, wallet.Bytes)

	bridgedTokens := groupTokens(bridgedInTokens)
	for _, bridgedOutToken := range groupTokens(bridgedOutTokens) {
		bridgedValue, err := getBridgedValue(bridgedTokens, bridgedOutToken.Identifier)
		require.Nil(t, err)

		fullTokenIdentifier := getTokenIdentifier(bridgedOutToken)
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, wallet.Bech32, big.NewInt(0).Sub(bridgedValue, bridgedOutToken.Amount))
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, esdtSafeEncoded, big.NewInt(0))

		tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(fullTokenIdentifier)
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, bridgedOutToken.Amount.String(), tokenSupply.Burned)
	}
}

func getIssuePaymentTokenInWallet(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	token string,
	amount *big.Int,
) {
	tokenData := &esdt.ESDigitalToken{
		Value: amount,
		Type:  uint32(core.Fungible),
	}
	marshalledTokenData, err := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer().Marshal(tokenData)
	require.NoError(t, err)

	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + token))
	tokenValue := hex.EncodeToString(marshalledTokenData)
	keyValueMap := map[string]string{
		tokenKey: tokenValue,
	}

	err = cs.SetKeyValueForAddress(wallet.Bech32, keyValueMap)
	require.NoError(t, err)
}

func getSovereignTokens(bridgedTokens []chainSim.ArgsDepositToken, prefix string) []string {
	tokens := make([]string, 0)
	for _, token := range bridgedTokens {
		if strings.HasPrefix(token.Identifier, prefix+"-") {
			tokens = append(tokens, token.Identifier)
		}
	}
	return tokens
}

func getBridgedValue(bridgeInTokens []chainSim.ArgsDepositToken, token string) (*big.Int, error) {
	for _, tkn := range bridgeInTokens {
		if tkn.Identifier == token {
			return tkn.Amount, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}
