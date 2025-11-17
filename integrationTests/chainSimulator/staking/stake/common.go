package stake

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	processChainSimulator "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	DefaultPathToInitialConfig = "../../../../cmd/node/config/"
)

var (
	RoundDurationInMillis          = uint64(6000)
	SupernovaRoundDurationInMillis = uint64(600)
	RoundsPerEpoch                 = core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	SupernovaRoundsPerEpoch = core.OptionalUint64{
		HasValue: true,
		Value:    200,
	}

	Log = logger.GetOrCreate("integrationTests/chainSimulator")
)

func TestBLSKeyStaked(t *testing.T,
	metachainNode processChainSimulator.NodeHandler,
	blsKey string,
) {
	decodedBLSKey, _ := hex.DecodeString(blsKey)
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	validatorStatistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		require.Equal(t, staking.StakedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
		return
	}

	// in staking ph 2/3.5 we do not find the bls key on the validator statistics
	_, found := validatorStatistics[blsKey]
	require.False(t, found)
	require.Equal(t, staking.QueuedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
}

func CheckExpectedStakedValue(t *testing.T, metachainNode processChainSimulator.NodeHandler, blsKey []byte, expectedValue int64) {
	totalStaked := GetTotalStaked(t, metachainNode, blsKey)

	expectedStaked := big.NewInt(expectedValue)
	expectedStaked = expectedStaked.Mul(chainSimulator.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(totalStaked))
}

func GetTotalStaked(t *testing.T, metachainNode processChainSimulator.NodeHandler, blsKey []byte) []byte {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulator.OkReturnCode, result.ReturnCode)

	return result.ReturnData[0]
}

func GetQualifiedAndUnqualifiedNodes(t *testing.T, metachainNode processChainSimulator.NodeHandler) ([]string, []string) {
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	qualified := make([]string, 0)
	unQualified := make([]string, 0)

	for _, auctionOwnerData := range auctionList {
		for _, auctionNode := range auctionOwnerData.Nodes {
			if auctionNode.Qualified {
				qualified = append(qualified, auctionNode.BlsKey)
			} else {
				unQualified = append(unQualified, auctionNode.BlsKey)
			}
		}
	}

	return qualified, unQualified
}

func GetUnStakedTokensList(t *testing.T, metachainNode processChainSimulator.NodeHandler, blsKey []byte) []byte {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulator.OkReturnCode, result.ReturnCode)

	return result.ReturnData[0]
}

func CheckOneOfTheNodesIsUnstaked(t *testing.T,
	metachainNode processChainSimulator.NodeHandler,
	blsKeys []string,
) {
	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	keyStatus0 := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	Log.Info("Key info", "key", blsKeys[0], "status", keyStatus0)

	isNotStaked0 := keyStatus0 == staking.UnStakedStatus

	decodedBLSKey1, _ := hex.DecodeString(blsKeys[1])
	keyStatus1 := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey1)
	Log.Info("Key info", "key", blsKeys[1], "status", keyStatus1)

	isNotStaked1 := keyStatus1 == staking.UnStakedStatus

	require.True(t, isNotStaked0 != isNotStaked1)
}
