package trieIterators

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type stakingValidatorInfo struct {
	totalStakedValue *big.Int
	topUpValue       *big.Int
}

type commonStakingProcessor struct {
	queryService process.SCQueryService
	blockChain   data.ChainHandler
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
	currentHeader := csp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return nil, ErrNodeNotInitialized
	}

	err := csp.accounts.RecreateTrie(currentHeader.GetRootHash())
	if err != nil {
		return nil, err
	}

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
