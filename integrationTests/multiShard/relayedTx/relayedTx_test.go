package relayedTx

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
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

		time.Sleep(time.Second)
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

	nodes, idxProposers, players, relayer, advertiser := CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
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
	scCode := arwen.GetSCCode("../../vm/arwen/testdata/erc20-c-03/wrc20_arwen.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		20000,
		make([]byte, 32),
		[]byte(arwen.CreateDeployTxData(scCode)+"@"+initialSupply),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	transferTokenVMGas := uint64(6960)
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

		time.Sleep(time.Second)
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

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithESDTTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	///////////------- send token issue
	tokenName := "robertWhyNot"
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	txData := "issue" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(initalSupply.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, issuePrice, vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, initalSupply)

	/////////------ send tx to players
	valueToTopUp := big.NewInt(100000000)
	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(valueToTopUp.Bytes())
	for _, player := range players {
		integrationTests.CreateAndSendTransaction(tokenIssuer, big.NewInt(0), player.Address, txData, integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(sendValue.Bytes())
	transferTokenESDTGas := uint64(1)
	transferTokenBaseGas := tokenIssuer.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte(txData)})
	transferTokenFullGas := transferTokenBaseGas + transferTokenESDTGas
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, big.NewInt(0), transferTokenFullGas, []byte(txData))
			_ = CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, big.NewInt(0), transferTokenFullGas, []byte(txData))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	checkAddressHasESDTTokens(t, receiverAddress1, nodes, tokenName, finalBalance)
	checkAddressHasESDTTokens(t, receiverAddress2, nodes, tokenName, finalBalance)

	players = append(players, relayer)
	checkPlayerBalances(t, nodes, players)
}

func TestRelayedTransactionInMultiShardEnvironmentWithAttestationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
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
	scCode := arwen.GetSCCode("attestation.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	registerValue := big.NewInt(100)
	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		200000,
		make([]byte, 32),
		[]byte(arwen.CreateDeployTxData(scCode)+"@"+hex.EncodeToString(registerValue.Bytes())+"@"+hex.EncodeToString(relayer.Address)+"@"+"ababab"),
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
	assert.Equal(t, actualBalance.Cmp(balance), 0)
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

func checkAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	userAcc := GetUserAccount(nodes, address)

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := getESDTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	assert.Equal(t, esdtData.Value.Cmp(value), 0)
}

func getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*builtInFunctions.ESDigitalToken, error) {
	esdtData := &builtInFunctions.ESDigitalToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		return esdtData, nil
	}

	err = integrationTests.TestMarshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}
