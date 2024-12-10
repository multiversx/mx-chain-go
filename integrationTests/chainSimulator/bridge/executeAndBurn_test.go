package bridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	dataApi "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
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

	nftV2 := sovChainPrefix + "-NFTV2-a1b2c3"
	nftV2Nonce := uint64(10)
	token := sovChainPrefix + "-TKN-1d2e3a"

	// TODO MX-15942 add dynamic NFT type for bridge transfer
	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: nftV2,
		Nonce:      nftV2Nonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: token,
		Nonce:      0,
		Amount:     big.NewInt(1),
		Type:       core.Fungible,
	})

	bridgedOutTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedOutTokens = append(bridgedOutTokens, chainSim.ArgsDepositToken{
		Identifier: nftV2,
		Nonce:      nftV2Nonce,
		Amount:     big.NewInt(1),
		Type:       core.NonFungibleV2,
	})

	simulateExecutionAndDeposit(t, bridgedInTokens, bridgedOutTokens)
}

func TestChainSimulator_ExecuteWithMintAndBurnSftWithDeposit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sft := sovChainPrefix + "-SOVSFT-654321"
	sftNonce := uint64(123)

	// TODO MX-15942 add dynamic SFT type for bridge transfer
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
			cfg.EpochConfig.EnableEpochs.DynamicESDTEnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(4)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(0)

	// Deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       sovChainPrefix,
		IssuePaymentToken: "WEGLD-bd4d79",
	}
	initOwnerAndSysAccState(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, argsEsdtSafe, enshrineEsdtSafeContract, enshrineEsdtSafeWasmPath)
	chainSim.RequireAccountHasToken(t, cs, argsEsdtSafe.IssuePaymentToken, initialAddress, big.NewInt(0))

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	// We need to register tokens originated from sovereign (to pay the issue cost)
	// Only the tokens with sovereign prefix need to be registered (these are the ones that will be minted), the rest will be taken from contract balance
	tokens := getUniquePrefixedTokens(bridgedInTokens, argsEsdtSafe.ChainPrefix)
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, tokens)

	// We will deposit an array of prefixed tokens from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our wallet address.
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, wallet.Bytes, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, nil)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	for _, bridgedInToken := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(bridgedInToken), wallet.Bech32, bridgedInToken.Amount)
		checkMetaDataInAccounts(t, cs, bridgedInToken, wallet.Bech32, bridgedInToken.Amount)
	}

	// deposit an array of tokens from main chain to sovereign chain,
	// expecting these tokens to be burned by the whitelisted ESDT safe sc
	txResult = deposit(t, cs, wallet.Bytes, &nonce, bridgeData.ESDTSafeAddress, bridgedOutTokens, wallet.Bytes)
	chainSim.RequireSuccessfulTransaction(t, txResult)

	bridgedTokens := groupTokens(bridgedInTokens)
	for _, bridgedOutToken := range groupTokens(bridgedOutTokens) {
		bridgedValue, err := getBridgedValue(bridgedTokens, bridgedOutToken.Identifier)
		require.Nil(t, err)

		fullTokenIdentifier := getTokenIdentifier(bridgedOutToken)
		remainingAmount := big.NewInt(0).Sub(bridgedValue, bridgedOutToken.Amount)
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, wallet.Bech32, remainingAmount)
		chainSim.RequireAccountHasToken(t, cs, fullTokenIdentifier, esdtSafeEncoded, big.NewInt(0))
		checkMetaDataInAccounts(t, cs, bridgedOutToken, wallet.Bech32, remainingAmount)

		tokenSupply, err := nodeHandler.GetFacadeHandler().GetTokenSupply(fullTokenIdentifier)
		require.Nil(t, err)
		require.NotNil(t, tokenSupply)
		require.Equal(t, bridgedOutToken.Amount.String(), tokenSupply.Burned)
	}
}

func getUniquePrefixedTokens(bridgedTokens []chainSim.ArgsDepositToken, prefix string) []string {
	tokens := make([]string, 0)
	seen := make(map[string]bool)
	for _, token := range bridgedTokens {
		if strings.HasPrefix(token.Identifier, prefix+"-") && !seen[token.Identifier] {
			tokens = append(tokens, token.Identifier)
			seen[token.Identifier] = true
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

func checkMetaDataInAccounts(
	t *testing.T,
	cs chainSim.ChainSimulator,
	token chainSim.ArgsDepositToken,
	account string,
	expectedAmount *big.Int,
) {
	addressShardID := chainSim.GetShardForAddress(cs, account)
	nodeHandler := cs.GetNodeHandler(addressShardID)

	// get user account token data
	esdtValue := getAccountTokenData(t, nodeHandler, account, token)

	if expectedAmount.Cmp(big.NewInt(0)) == 0 {
		require.Empty(t, esdtValue)            // expect that key doesn't exist in account
		if metaDataOnUserAccount(token.Type) { // for other token types the key can exist because other wallets in shard can have the token
			requireNoTokenDataInSysAccount(t, nodeHandler, token) // no keys in system account
		}
		return
	}

	esdtData := &esdt.ESDigitalToken{}
	err := nodeHandler.GetCoreComponents().InternalMarshalizer().Unmarshal(esdtData, esdtValue)
	require.Nil(t, err)
	require.NotNil(t, esdtData)
	require.Equal(t, expectedAmount, esdtData.Value)
	require.Equal(t, uint32(token.Type), esdtData.Type)

	if token.Type == core.Fungible {
		require.Nil(t, esdtData.TokenMetaData)
		requireNoTokenDataInSysAccount(t, nodeHandler, token) // no keys in system account
	} else if token.Type == core.NonFungibleV2 || token.Type == core.DynamicNFT {
		require.NotNil(t, esdtData.TokenMetaData)
		require.Equal(t, token.Nonce, esdtData.TokenMetaData.Nonce)
		requireNoTokenDataInSysAccount(t, nodeHandler, token) // no keys in system account
	} else {
		require.Nil(t, esdtData.TokenMetaData)

		// get system account token data
		esdtValue = getAccountTokenData(t, nodeHandler, chainSim.ESDTSystemAccount, token)

		esdtData = &esdt.ESDigitalToken{}
		err = nodeHandler.GetCoreComponents().InternalMarshalizer().Unmarshal(esdtData, esdtValue)
		require.Nil(t, err)
		require.NotNil(t, esdtData)
		require.GreaterOrEqual(t, esdtData.Value.Uint64(), expectedAmount.Uint64()) // greater if other wallets have the token
		require.Equal(t, uint32(token.Type), esdtData.Type)
		require.NotNil(t, esdtData.TokenMetaData)
		require.Equal(t, token.Nonce, esdtData.TokenMetaData.Nonce)
	}
}

func getAccountTokenData(
	t *testing.T,
	nodeHandler process.NodeHandler,
	account string,
	token chainSim.ArgsDepositToken,
) []byte {
	accountKeys, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(account, dataApi.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountKeys)

	esdtValue, err := hex.DecodeString(accountKeys[getTokenKey(token)])
	require.Nil(t, err)

	return esdtValue
}

func requireNoTokenDataInSysAccount(
	t *testing.T,
	nodeHandler process.NodeHandler,
	token chainSim.ArgsDepositToken,
) {
	accountKeys, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(chainSim.ESDTSystemAccount, dataApi.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, accountKeys)
	require.Empty(t, accountKeys[getTokenKey(token)])
}

func metaDataOnUserAccount(esdtType core.ESDTType) bool {
	return esdtType == core.Fungible ||
		esdtType == core.NonFungibleV2 ||
		esdtType == core.DynamicNFT
}

func getTokenKey(token chainSim.ArgsDepositToken) string {
	nonce := ""
	if token.Nonce > 0 {
		nonce = hex.EncodeToString(big.NewInt(0).SetUint64(token.Nonce).Bytes())
	}
	return hex.EncodeToString([]byte(core.ProtectedKeyPrefix+core.ESDTKeyIdentifier)) +
		hex.EncodeToString([]byte(token.Identifier)) +
		nonce
}
