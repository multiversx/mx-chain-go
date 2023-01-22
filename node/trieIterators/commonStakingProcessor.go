package trieIterators

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type stakingValidatorInfo struct {
	totalStakedValue *big.Int
	topUpValue       *big.Int
}

type commonStakingProcessor struct {
	queryService process.SCQueryService
	accounts     *AccountsWrapper
}

func (csp *commonStakingProcessor) getValidatorInfoFromSC(validatorAddress []byte) (*stakingValidatorInfo, error) {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStakedTopUpStakedBlsKeys",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorAddress},
	}

	vmOutput, err := csp.queryService.ExecuteQuery(scQuery)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, return code: %v, message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	if len(vmOutput.ReturnData) < 3 {
		return nil, fmt.Errorf("%w, getTotalStakedTopUpStakedBlsKeys function should have at least three values", epochStart.ErrExecutingSystemScCode)
	}

	info := &stakingValidatorInfo{}

	info.topUpValue = big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	info.totalStakedValue = big.NewInt(0).SetBytes(vmOutput.ReturnData[1])

	return info, nil
}

func (csp *commonStakingProcessor) getAccount(scAddress []byte) (state.UserAccountHandler, error) {
	accountHandler, err := csp.accounts.GetExistingAccount(scAddress)
	if err != nil {
		return nil, err
	}

	account, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return nil, ErrCannotCastAccountHandlerToUserAccount
	}

	return account, nil
}
