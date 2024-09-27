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
	"github.com/multiversx/mx-chain-go/process"
)

const (
	saveCallerWasmPath = "testdata/tc-contract.wasm"
)

// Test steps:
// - deploy sovereign bridge contracts on the main chain
// - whitelist the bridge esdt safe contract to allow it to burn/mint cross chain esdt tokens
// - deploy a receiver contract for transfer and execute testing
// - executeOperation token with prefix and contract call
// - verify operation was executed in the name of user
func TestChainSimulator_ExecuteOperationWithTransferDataForTokenWithPrefix(t *testing.T) {
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
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	// Deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       "sov1",
		IssuePaymentToken: "ABC-123456",
	}
	setStateBridgeOwner(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, esdtSafeWasmPath, argsEsdtSafe, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	// Create a new user account
	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	// Deploy a receiver contract for transfer and execute testing
	taeContractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", saveCallerWasmPath)
	taeContractAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(taeContractAddress)

	tokenToBridge := argsEsdtSafe.ChainPrefix + "-SOVT-5d8f56"
	expectedMintValue := big.NewInt(123)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: tokenToBridge,
		Nonce:      0,
		Amount:     expectedMintValue,
		Type:       core.Fungible,
	})

	// Register the sovereign token in contract
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, []string{tokenToBridge})

	scTransferData := &transferData{
		GasLimit: uint64(50000000),
		Function: []byte("pay"),
		Args:     [][]byte{},
	}

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc,
	// bridge service wallet makes the executeBridgeOps
	// tae contract is the receiver and called on "pay" endpoint (just saves the tokens in balance and updates a storage with last caller)
	// expecting tokens will be in tae contract
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, taeContractAddress, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, scTransferData)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	for _, token := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), taeContractAddressBech32, token.Amount)
	}

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  taeContractAddress,
		FuncName:   "getLastPayCaller",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)
	lastCaller := res.ReturnData[0]
	require.Equal(t, wallet.Bytes, lastCaller)
}

// Test steps:
// - deploy sovereign bridge contracts on the main chain
// - whitelist the bridge esdt safe contract to allow it to burn/mint cross chain esdt tokens
// - deploy a receiver contract for transfer and execute testing
// - executeOperation token with prefix and contract call
// - verify operation was executed in the name of user
func TestChainSimulator_ExecuteOperationWithFailedTransferDataForTokenWithPrefix(t *testing.T) {
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
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	// Deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       "sov1",
		IssuePaymentToken: "ABC-123456",
	}
	setStateBridgeOwner(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, esdtSafeWasmPath, argsEsdtSafe, feeMarketWasmPath)

	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.Equal(t, whiteListedAddress, esdtSafeEncoded)

	// Create a new user account
	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	// Deploy a receiver contract for transfer and execute testing
	taeContractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", saveCallerWasmPath)

	tokenToBridge := argsEsdtSafe.ChainPrefix + "-SOVT-5d8f56"
	expectedMintValue := big.NewInt(123)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: tokenToBridge,
		Nonce:      0,
		Amount:     expectedMintValue,
		Type:       core.Fungible,
	})

	// Register the sovereign token in contract
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, []string{tokenToBridge})

	scTransferData := &transferData{
		GasLimit: uint64(50000000),
		Function: []byte("payy"),
		Args:     [][]byte{},
	}

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc,
	// bridge service wallet makes the executeBridgeOps
	// tae contract is the receiver and called on "payy" endpoint which doesn't exist
	// expecting tokens to be returned in user wallet
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, taeContractAddress, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, scTransferData)
	chainSim.RequireErrorEvent(t, txResult, "internalVMErrors", "invalid function (not found)")
	for _, token := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), wallet.Bech32, token.Amount)
	}

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  taeContractAddress,
		FuncName:   "getLastPayCaller",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	require.Equal(t, "storage decode error: bad array length", res.ReturnMessage) // storage is empty because endpoint was not called
}

// Test steps:
// - deploy sovereign bridge contracts on the main chain
// - whitelist the bridge esdt safe contract to allow it to burn/mint cross chain esdt tokens
// - deploy a receiver contract for transfer and execute testing
// - executeOperation token with prefix and contract call
// - verify operation was executed in the name of user
func TestChainSimulator_ExecuteOperationWithTransferDataForTokenWithPrefixContractNotWhitelisted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqcw92wj0huvaghg4aeuykknp7hstmrmhudz3shjdhtt"
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
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	// Deploy bridge setup
	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	argsEsdtSafe := ArgsEsdtSafe{
		ChainPrefix:       "sov1",
		IssuePaymentToken: "ABC-123456",
	}
	setStateBridgeOwner(t, cs, initialAddress, argsEsdtSafe)
	bridgeData := deployBridgeSetup(t, cs, initialAddress, esdtSafeWasmPath, argsEsdtSafe, feeMarketWasmPath)

	// esdt-safe address generated is NOT whitelisted
	esdtSafeEncoded, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(bridgeData.ESDTSafeAddress)
	require.NotEqual(t, whiteListedAddress, esdtSafeEncoded)

	// Create a new user account
	wallet, err := cs.GenerateAndMintWalletAddress(0, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)
	paymentTokenAmount, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	chainSim.SetEsdtInWallet(t, cs, wallet, argsEsdtSafe.IssuePaymentToken, 0, esdt.ESDigitalToken{Value: paymentTokenAmount})

	// Deploy a receiver contract for transfer and execute testing
	taeContractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", saveCallerWasmPath)
	taeContractAddressBech32, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(taeContractAddress)

	tokenToBridge := argsEsdtSafe.ChainPrefix + "-SOVT-5d8f56"
	expectedMintValue := big.NewInt(123)

	bridgedInTokens := make([]chainSim.ArgsDepositToken, 0)
	bridgedInTokens = append(bridgedInTokens, chainSim.ArgsDepositToken{
		Identifier: tokenToBridge,
		Nonce:      0,
		Amount:     expectedMintValue,
		Type:       core.Fungible,
	})

	// Register the sovereign token in contract
	registerSovereignNewTokens(t, cs, wallet, &nonce, bridgeData.ESDTSafeAddress, argsEsdtSafe.IssuePaymentToken, []string{tokenToBridge})

	scTransferData := &transferData{
		GasLimit: uint64(50000000),
		Function: []byte("pay"),
		Args:     [][]byte{},
	}

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc,
	// bridge service wallet makes the executeBridgeOps
	// tae contract is the receiver and called on "pay" endpoint (just saves the tokens in balance and updates a storage with last caller)
	// expecting tokens will be in tae contract

	//should call another endpoint because this will give action not allowed for mint
	txResult := executeOperation(t, cs, bridgeData.OwnerAccount.Wallet, taeContractAddress, &bridgeData.OwnerAccount.Nonce, bridgeData.ESDTSafeAddress, bridgedInTokens, wallet.Bytes, scTransferData)
	chainSim.RequireSuccessfulTransaction(t, txResult)
	for _, token := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), taeContractAddressBech32, token.Amount)
	}

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  taeContractAddress,
		FuncName:   "getLastPayCaller",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)
	lastCaller := res.ReturnData[0]
	require.Equal(t, bridgeData.ESDTSafeAddress, lastCaller) // EXPECTED
}
