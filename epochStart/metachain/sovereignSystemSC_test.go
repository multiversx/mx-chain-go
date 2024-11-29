package metachain

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-go/process/mock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	systemVm "github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestSovereignSystemSC_ProcessDelegationRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgsForSystemSCProcessor()

	receiver := []byte("rcv")
	rewardVal1 := big.NewInt(100)
	rewardVal2 := big.NewInt(101)

	balanceAddedCt := 0
	userAcc := &stateMock.UserAccountStub{
		AddToBalanceCalled: func(value *big.Int) error {
			balanceAddedCt++
			require.Equal(t, rewardVal2, value)
			return nil
		},
	}

	saveAccCt := 0
	args.UserAccountsDB = &stateMock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			require.Equal(t, receiver, container)
			return userAcc, nil
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			saveAccCt++
			require.Equal(t, userAcc, account)
			return nil
		},
	}

	sysAccOnMetaUpdatedCt := 0
	args.SystemVM = &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			require.Equal(t, &vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr: args.EndOfEpochCallerAddress,
					CallValue:  rewardVal1,
				},
				RecipientAddr: systemVm.StakingSCAddress,
				Function:      "updateRewards",
			}, input)

			sysAccOnMetaUpdatedCt++
			return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
		},
	}

	sysSC, _ := NewSystemSCProcessor(args)
	sovSysSC, err := NewSovereignSystemSCProcessor(sysSC)
	require.Nil(t, err)
	require.False(t, sovSysSC.IsInterfaceNil())

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHash3 := []byte("txHash3")
	mbs := block.MiniBlockSlice{
		{
			TxHashes: [][]byte{txHash1, txHash3},
			Type:     block.RewardsBlock,
		},
		{
			TxHashes: [][]byte{txHash2},
			Type:     block.PeerBlock,
		},
	}

	tx1 := &rewardTx.RewardTx{
		RcvAddr: systemVm.StakingSCAddress,
		Value:   rewardVal1,
	}
	tx2 := &rewardTx.RewardTx{
		RcvAddr: receiver,
		Value:   rewardVal2,
	}

	txCache := &mock.TxForCurrentBlockStub{
		GetTxCalled: func(txHash []byte) (data.TransactionHandler, error) {
			switch string(txHash) {
			case string(txHash1):
				return tx1, nil
			case string(txHash3):
				return tx2, nil
			}

			return nil, errors.New("invalid tx hash")
		},
	}

	err = sovSysSC.ProcessDelegationRewards(mbs, txCache)
	require.Nil(t, err)
	require.Equal(t, 1, balanceAddedCt)
	require.Equal(t, 1, saveAccCt)
	require.Equal(t, 1, sysAccOnMetaUpdatedCt)
}
