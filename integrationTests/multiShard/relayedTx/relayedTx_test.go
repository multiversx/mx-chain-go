package relayedTx

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer := CreateGeneralSetupForRelayTxTest()
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(integrationTests.StepDelay)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := GetUserAccount(nodes, receiverAddress1)
	receiver2 := GetUserAccount(nodes, receiverAddress2)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	assert.Equal(t, receiver1.GetBalance().Cmp(finalBalance), 0)
	assert.Equal(t, receiver2.GetBalance().Cmp(finalBalance), 0)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithSmartContractTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer := CreateGeneralSetupForRelayTxTest()
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	ownerNode := nodes[0]
	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	scCode := wasm.GetSCCode("../../vm/wasm/testdata/erc20-c-03/wrc20_wasm.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		20000,
		make([]byte, 32),
		[]byte(wasm.CreateDeployTxData(scCode)+"@"+initialSupply),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	transferTokenVMGas := uint64(7200)
	transferTokenBaseGas := ownerNode.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte("transferToken@" + hex.EncodeToString(receiverAddress1) + "@00" + hex.EncodeToString(sendValue.Bytes()))})
	transferTokenFullGas := transferTokenBaseGas + transferTokenVMGas

	initialTokenSupply := big.NewInt(1000000000)
	initialPlusForGas := uint64(1000)
	for _, player := range players {
		integrationTests.CreateAndSendTransactionWithGasLimit(
			ownerNode,
			big.NewInt(0),
			transferTokenFullGas+initialPlusForGas,
			scAddress,
			[]byte("transferToken@"+hex.EncodeToString(player.Address)+"@00"+hex.EncodeToString(initialTokenSupply.Bytes())),
			integrationTests.ChainID,
			integrationTests.MinTransactionVersion,
		)
	}
	time.Sleep(time.Second)

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		for _, player := range players {
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress1)+"@00"+hex.EncodeToString(sendValue.Bytes())))
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress2)+"@00"+hex.EncodeToString(sendValue.Bytes())))
		}

		time.Sleep(integrationTests.StepDelay)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)

	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)
	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)

	checkPlayerBalances(t, nodes, players)

	userAcc := GetUserAccount(nodes, relayer.Address)
	assert.Equal(t, userAcc.GetBalance().Cmp(relayer.Balance), 1)
}

func TestRelayedTransactionInMultiShardEnvironmentWithESDTTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer := CreateGeneralSetupForRelayTxTest()
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	// ------- send token issue
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	txData := "issue" +
		"@" + hex.EncodeToString([]byte("robertWhyNot")) +
		"@" + hex.EncodeToString([]byte("RBT")) +
		"@" + hex.EncodeToString(initalSupply.Bytes()) +
		"@" + hex.EncodeToString([]byte{6})
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(integrationTests.StepDelay)
	}
	time.Sleep(time.Second)

	tokenIdenfitifer := string(integrationTests.GetTokenIdentifier(nodes, []byte("RBT")))
	CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdenfitifer, initalSupply)

	// ------ send tx to players
	valueToTopUp := big.NewInt(100000000)
	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenIdenfitifer)) + "@" + hex.EncodeToString(valueToTopUp.Bytes())
	for _, player := range players {
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), player.Address, txData, integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(integrationTests.StepDelay)
	}
	time.Sleep(time.Second)

	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenIdenfitifer)) + "@" + hex.EncodeToString(sendValue.Bytes())
	transferTokenESDTGas := uint64(1)
	transferTokenBaseGas := tokenIssuer.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte(txData)})
	transferTokenFullGas := transferTokenBaseGas + transferTokenESDTGas + uint64(100) // use more gas to simulate gas refund
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, big.NewInt(0), transferTokenFullGas, []byte(txData))
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, big.NewInt(0), transferTokenFullGas, []byte(txData))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(integrationTests.StepDelay)
	}

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	CheckAddressHasTokens(t, receiverAddress1, nodes, tokenIdenfitifer, finalBalance)
	CheckAddressHasTokens(t, receiverAddress2, nodes, tokenIdenfitifer, finalBalance)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithAttestationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer := CreateGeneralSetupForRelayTxTest()
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	for _, node := range nodes {
		node.EconomicsData.SetMaxGasLimitPerBlock(1500000000)
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	ownerNode := nodes[0]
	scCode := wasm.GetSCCode("attestation.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	registerValue := big.NewInt(100)
	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		200000,
		make([]byte, 32),
		[]byte(wasm.CreateDeployTxData(scCode)+"@"+hex.EncodeToString(registerValue.Bytes())+"@"+hex.EncodeToString(relayer.Address)+"@"+"ababab"),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	time.Sleep(time.Second)

	registerVMGas := uint64(100000)
	savePublicInfoVMGas := uint64(100000)
	attestVMGas := uint64(100000)

	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

	uniqueIDs := make([]string, len(players))
	for i, player := range players {
		uniqueIDs[i] = core.UniqueIdentifier()
		_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, registerValue,
			registerVMGas, []byte("register@"+hex.EncodeToString([]byte(uniqueIDs[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	cryptoHook := hooks.NewVMCryptoHook()
	privateInfos := make([]string, len(players))
	for i := range players {
		privateInfos[i] = core.UniqueIdentifier()
		publicInfo, _ := cryptoHook.Keccak256([]byte(privateInfos[i]))
		createAndSendSimpleTransaction(nodes, relayer, scAddress, big.NewInt(0), savePublicInfoVMGas,
			[]byte("savePublicInfo@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString(publicInfo)))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagate := int64(5)
	for i := int64(0); i <= nrRoundsToPropagate; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0), attestVMGas,
			[]byte("attest@"+hex.EncodeToString([]byte(uniqueIDs[i]))+"@"+hex.EncodeToString([]byte(privateInfos[i]))))
		_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, registerValue,
			registerVMGas, []byte("register@"+hex.EncodeToString([]byte(uniqueIDs[i]))))
	}
	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	for i, player := range players {
		checkAttestedPublicKeys(t, ownerNode, scAddress, []byte(uniqueIDs[i]), player.Address)
	}
}

func checkAttestedPublicKeys(
	t *testing.T,
	node *integrationTests.TestProcessorNode,
	scAddress []byte,
	obfuscatedData []byte,
	userAddress []byte,
) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "getPublicKey",
		Arguments: [][]byte{obfuscatedData},
	})
	require.Nil(t, err)
	require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
	require.Equal(t, vmOutput.ReturnData[0], userAddress)
}

func checkSCBalance(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, userAddress []byte, balance *big.Int) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "balanceOf",
		Arguments: [][]byte{userAddress},
	})
	assert.Nil(t, err)
	actualBalance := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	assert.Equal(t, 0, actualBalance.Cmp(balance))
}

func checkPlayerBalances(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount) {
	for _, player := range players {
		userAcc := GetUserAccount(nodes, player.Address)
		assert.Equal(t, userAcc.GetBalance().Cmp(player.Balance), 0)
		assert.Equal(t, userAcc.GetNonce(), player.Nonce)
	}
}

func CheckAddressHasTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	userAcc := GetUserAccount(nodes, address)

	tokenKey := []byte(core.ProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := getESDTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	assert.Equal(t, esdtData.Value.Cmp(value), 0)
}

func getESDTDataFromKey(userAcnt common.UserAccountHandler, key []byte) (*esdt.ESDigitalToken, error) {
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0)}
	marshaledData, _, err := userAcnt.RetrieveValue(key)
	if err != nil {
		return esdtData, nil
	}

	err = integrationTests.TestMarshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}
