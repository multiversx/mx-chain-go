package intermediate

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewDelegationProcessor_NilExecutorShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		nil,
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{},
		&mock.NodesListSplitterStub{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilTxExecutionProcessor, err)
}

func TestNewDelegationProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{},
		nil,
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{},
		&mock.NodesListSplitterStub{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestNewDelegationProcessor_NilAccountsParserShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{},
		&mock.ShardCoordinatorMock{},
		nil,
		&mock.SmartContractParserStub{},
		&mock.NodesListSplitterStub{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilAccountsParser, err)
}

func TestNewDelegationProcessor_NilSmartContractParserShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{},
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		nil,
		&mock.NodesListSplitterStub{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilSmartContractParser, err)
}

func TestNewDelegationProcessor_NilNodesSplitterShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{},
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{},
		nil,
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilNodesListSplitter, err)
}

func TestNewDelegationProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	dp, err := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{},
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{},
		&mock.NodesListSplitterStub{},
	)

	assert.False(t, check.IfNil(dp))
	assert.Nil(t, err)
}

//------- ExecuteDelegation

func TestDelegationProcessor_ExecuteDelegationSplitFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	dp, _ := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{
			ExecuteTransactionCalled: func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
				assert.Fail(t, "should have not execute a transaction")

				return nil
			},
		},
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{
			InitialSmartContractsSplitOnOwnersShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialSmartContractHandler, error) {
				return nil, expectedErr
			},
		},
		&mock.NodesListSplitterStub{},
	)

	result, err := dp.ExecuteDelegation()

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, genesis.DelegationResult{}, result)
}

func TestDelegationProcessor_ExecuteDelegationNoDelegationScShouldRetNil(t *testing.T) {
	t.Parallel()

	dp, _ := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{
			ExecuteTransactionCalled: func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
				assert.Fail(t, "should have not execute a transaction")

				return nil
			},
		},
		&mock.ShardCoordinatorMock{},
		&mock.AccountsParserStub{},
		&mock.SmartContractParserStub{
			InitialSmartContractsSplitOnOwnersShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialSmartContractHandler, error) {
				return map[uint32][]genesis.InitialSmartContractHandler{
					0: {
						&data.InitialSmartContract{
							Type: "test",
						},
					},
				}, nil
			},
		},
		&mock.NodesListSplitterStub{},
	)

	result, err := dp.ExecuteDelegation()

	assert.Nil(t, err)
	assert.Equal(t, genesis.DelegationResult{}, result)
}

func TestDelegationProcessor_ExecuteDelegationStakeShouldWork(t *testing.T) {
	t.Parallel()

	staker1 := []byte("stakerB")
	staker2 := []byte("stakerC")
	delegationSc := []byte("delegation SC")

	dp, _ := NewDelegationProcessor(
		&mock.TxExecutionProcessorStub{
			ExecuteTransactionCalled: func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
				isStakeCall := strings.Contains(string(data), "stake")
				isStaker := bytes.Equal(sndAddr, staker1) || bytes.Equal(sndAddr, staker2)
				if isStakeCall && !isStaker {
					assert.Fail(t, "stake should have been called by the one of the stakers")
				}

				return nil
			},
		},
		&mock.ShardCoordinatorMock{
			SelfShardId: 0,
			NumOfShards: 2,
		},
		&mock.AccountsParserStub{
			GetInitialAccountsForDelegatedCalled: func(addressBytes []byte) []genesis.InitialAccountHandler {
				if bytes.Equal(addressBytes, delegationSc) {
					ia1 := &data.InitialAccount{
						Delegation: &data.DelegationData{},
					}
					ia1.SetAddressBytes(staker1)

					ia2 := &data.InitialAccount{
						Delegation: &data.DelegationData{},
					}
					ia2.SetAddressBytes(staker2)

					return []genesis.InitialAccountHandler{ia1, ia2}
				}

				return make([]genesis.InitialAccountHandler, 0)
			},
		},
		&mock.SmartContractParserStub{
			InitialSmartContractsSplitOnOwnersShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialSmartContractHandler, error) {
				sc := &data.InitialSmartContract{
					Type: genesis.DelegationType,
				}
				sc.SetAddressBytes(delegationSc)

				return map[uint32][]genesis.InitialSmartContractHandler{
					0: {sc},
				}, nil
			},
		},
		&mock.NodesListSplitterStub{},
	)

	result, err := dp.ExecuteDelegation()

	expectedResult := genesis.DelegationResult{
		NumTotalDelegated: 0,
		NumTotalStaked:    2,
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, result)
}
