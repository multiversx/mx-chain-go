package staking

import (
	"encoding/hex"
	"math/big"
	"testing"

	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	// MockBLSSignature the const for a mocked bls signature
	MockBLSSignature = "010101"
	// GasLimitForStakeOperation the const for the gas limit value for the stake operation
	GasLimitForStakeOperation = 50_000_000
	// GasLimitForUnBond the const for the gas limit value for the unBond operation
	GasLimitForUnBond = 12_000_000
	// MaxNumOfBlockToGenerateWhenExecutingTx the const for the maximum number of block to generate when execute a transaction
	MaxNumOfBlockToGenerateWhenExecutingTx = 7

	// QueuedStatus the const for the queued status of a validators
	QueuedStatus = "queued"
	// StakedStatus the const for the staked status of a validators
	StakedStatus = "staked"
	// NotStakedStatus the const for the notStaked status of a validators
	NotStakedStatus = "notStaked"
	// AuctionStatus the const for the action status of a validators
	AuctionStatus = "auction"
	// UnStakedStatus the const for the unStaked status of a validators
	UnStakedStatus = "unStaked"
)

var (
	//InitialDelegationValue the variable for the initial delegation value
	InitialDelegationValue = big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(1250))
)

// GetNonce will return the nonce of the provided address
func GetNonce(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, address dtos.WalletAddress) uint64 {
	account, err := cs.GetAccount(address)
	require.Nil(t, err)

	return account.Nonce
}

// GetBLSKeyStatus will return the bls key status
func GetBLSKeyStatus(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) string {
	scQuery := &process.SCQuery{
		ScAddress:  vm.StakingSCAddress,
		FuncName:   "getBLSKeyStatus",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	return string(result.ReturnData[0])
}

// GetAllNodeStates will return the status of all the nodes that belong to the provided address
func GetAllNodeStates(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte) map[string]string {
	scQuery := &process.SCQuery{
		ScAddress:  address,
		FuncName:   "getAllNodeStates",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	m := make(map[string]string)
	status := ""
	for _, resultData := range result.ReturnData {
		if len(resultData) != 96 {
			// not a BLS key
			status = string(resultData)
			continue
		}

		m[hex.EncodeToString(resultData)] = status
	}

	return m
}

// CheckValidatorStatus will compare the status of the provided bls key with the provided expected status
func CheckValidatorStatus(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, blsKey string, expectedStatus string) {
	err := cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	validatorsStatistics, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	require.Equal(t, expectedStatus, validatorsStatistics[blsKey].ValidatorStatus)
}
