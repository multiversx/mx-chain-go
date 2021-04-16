package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testVm "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/require"
)

// GetESDTTokenData -
func GetESDTTokenData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
) *esdt.ESDigitalToken {
	userAcc := GetUserAccountWithAddress(t, address, nodes)
	require.False(t, check.IfNil(userAcc))

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := GetESDTDataFromKey(userAcc, tokenKey)
	require.Nil(t, err)

	return esdtData
}

// GetUserAccountWithAddress -
func GetUserAccountWithAddress(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
) state.UserAccountHandler {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				acc, err := helperNode.AccntState.LoadAccount(address)
				require.Nil(t, err)
				return acc.(state.UserAccountHandler)
			}
		}
	}

	return nil
}

// GetESDTDataFromKey -
func GetESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*esdt.ESDigitalToken, error) {
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0)}
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

// SetRoles -
func SetRoles(nodes []*integrationTests.TestProcessorNode, addrForRole []byte, tokenIdentifier []byte, roles [][]byte) {
	tokenIssuer := nodes[0]

	txData := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(addrForRole)

	for _, role := range roles {
		txData += "@" + hex.EncodeToString(role)
	}

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)
}

// DeployNonPayableSmartContract -
func DeployNonPayableSmartContract(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nonce *uint64,
	round *uint64,
	fileName string,
) []byte {
	// deploy Smart Contract which can do local mint and local burn
	scCode := arwen.GetSCCode(fileName)
	scAddress, _ := nodes[0].BlockchainHook.NewAddress(nodes[0].OwnAccount.Address, nodes[0].OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxDataNonPayable(scCode),
		integrationTests.AdditionalGasLimit,
	)

	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, 4, *nonce, *round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	return scAddress
}

// CheckAddressHasESDTTokens -
func CheckAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value int64,
) {
	esdtData := GetESDTTokenData(t, address, nodes, tokenName)
	bigValue := big.NewInt(value)
	if esdtData.Value.Cmp(bigValue) != 0 {
		require.Fail(t, fmt.Sprintf("esdt balance difference. expected %s, but got %s", bigValue.String(), esdtData.Value.String()))
	}
}

// CreateNodesAndPrepareBalances -
func CreateNodesAndPrepareBalances(numOfShards int) ([]*integrationTests.TestProcessorNode, []int) {
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard
	integrationTests.DisplayAndStartNodes(nodes)

	return nodes, idxProposers
}

// IssueNFT -
func IssueNFT(nodes []*integrationTests.TestProcessorNode, esdtType string, ticker string) {
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]

	txData := txDataBuilder.NewBuilder()

	issueFunc := "issueNonFungible"
	if esdtType == core.SemiFungibleESDT {
		issueFunc = "issueSemiFungible"
	}
	txData.Clear().Func(issueFunc).Str(tokenName).Str(ticker)
	txData.CanFreeze(false).CanWipe(false).CanPause(false).CanTransferNFTCreateRole(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)
}

// IssueTestToken -
func IssueTestToken(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string) {
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]

	txData := txDataBuilder.NewBuilder()

	txData.Clear().IssueESDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)
}

// CheckNumCallBacks -
func CheckNumCallBacks(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	expectedNumCallbacks int,
) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
		require.Equal(t, expectedNumCallbacks, len(vmOutput.ReturnData))
	}
}

// CheckSavedCallBackData -
func CheckSavedCallBackData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	callbackIndex int,
	expectedTokenId string,
	expectedPayment *big.Int,
	expectedResultCode vmcommon.ReturnCode,
	expectedArguments [][]byte) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data_at_index",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments: [][]byte{
				{byte(callbackIndex)},
			},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
		require.GreaterOrEqual(t, 3, len(vmOutput.ReturnData))
		require.Equal(t, []byte(expectedTokenId), vmOutput.ReturnData[0])
		require.Equal(t, expectedPayment.Bytes(), vmOutput.ReturnData[1])
		if expectedResultCode == vmcommon.Ok {
			require.Equal(t, []byte{}, vmOutput.ReturnData[2])
		} else {
			require.Equal(t, []byte{byte(expectedResultCode)}, vmOutput.ReturnData[2])
		}
		require.Equal(t, expectedArguments, vmOutput.ReturnData[3:])
	}
}

// PrepareFungibleTokensWithLocalBurnAndMint -
func PrepareFungibleTokensWithLocalBurnAndMint(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	addressWithRoles []byte,
	idxProposers []int,
	round *uint64,
	nonce *uint64,
) string {
	IssueTestToken(nodes, 100, "TKN")

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKN")))

	SetRoles(nodes, addressWithRoles, []byte(tokenIdentifier), [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleLocalBurn)})

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	return tokenIdentifier
}
