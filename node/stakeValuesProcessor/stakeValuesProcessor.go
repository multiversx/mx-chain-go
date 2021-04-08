package stakeValuesProcessor

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type stakedValuesProc struct {
	accounts                        state.AccountsAdapter
	blockChain                      data.ChainHandler
	queryService                    process.SCQueryService
	publicKeyConverter              core.PubkeyConverter
	mutexComputeStakedValueAndTopUp *sync.Mutex
}

// NewTotalStakedValueProcessor will create a new instance of stakedValuesProc
func NewTotalStakedValueProcessor(
	accounts state.AccountsAdapter,
	blockChain data.ChainHandler,
	queryService process.SCQueryService,
	publicKeyConverter core.PubkeyConverter,
) (*stakedValuesProc, error) {
	if check.IfNil(accounts) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(blockChain) {
		return nil, ErrNilBlockChain
	}
	if check.IfNil(queryService) {
		return nil, ErrNilQueryService
	}
	if check.IfNil(publicKeyConverter) {
		return nil, ErrNilPubkeyConverter
	}

	return &stakedValuesProc{
		accounts:                        accounts,
		blockChain:                      blockChain,
		mutexComputeStakedValueAndTopUp: &sync.Mutex{},
		queryService:                    queryService,
		publicKeyConverter:              publicKeyConverter,
	}, nil
}

// GetTotalStakedValue will calculate total staked value if needed and return calculated value
func (svp *stakedValuesProc) GetTotalStakedValue() (*api.StakeValues, error) {
	totalStaked, topUp, err := svp.computeStakedValueAndTopUp()
	if err != nil {
		return nil, err
	}

	return &api.StakeValues{
		TotalStaked: totalStaked,
		TopUp:       topUp,
	}, nil
}

func (svp *stakedValuesProc) computeStakedValueAndTopUp() (*big.Int, *big.Int, error) {
	svp.mutexComputeStakedValueAndTopUp.Lock()
	defer svp.mutexComputeStakedValueAndTopUp.Unlock()

	currentHeader := svp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return nil, nil, nil
	}

	err := svp.accounts.RecreateTrie(currentHeader.GetRootHash())
	if err != nil {
		return nil, nil, err
	}

	ah, err := svp.accounts.GetExistingAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, nil, err
	}

	account, ok := ah.(state.UserAccountHandler)
	if !ok {
		return nil, nil, ErrCannotCastAccountHandlerToUserAccount
	}

	rootHash, err := account.DataTrie().RootHash()
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	chLeaves, err := account.DataTrie().GetAllLeavesOnChannel(rootHash, ctx)
	if err != nil {
		return nil, nil, err
	}

	totalStaked, totalTopUp := big.NewInt(0), big.NewInt(0)
	for leaf := range chLeaves {
		leafKey := leaf.Key()
		if len(leafKey) != svp.publicKeyConverter.Len() {
			continue
		}

		totalStakedCurrentAccount, totalTopUpCurrentAccount, errGet := svp.getValidatorInfoFromSC(leafKey)
		if errGet != nil {
			continue
		}

		totalStaked = totalStaked.Add(totalStaked, totalStakedCurrentAccount)
		totalTopUp = totalTopUp.Add(totalTopUp, totalTopUpCurrentAccount)
	}

	return totalStaked, totalTopUp, nil
}

func (svp *stakedValuesProc) getValidatorInfoFromSC(validatorAddress []byte) (*big.Int, *big.Int, error) {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStakedTopUpStakedBlsKeys",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorAddress},
	}

	vmOutput, err := svp.queryService.ExecuteQuery(scQuery)
	if err != nil {
		return nil, nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, nil, fmt.Errorf("%w, error: %v message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	if len(vmOutput.ReturnData) < 3 {
		return nil, nil, fmt.Errorf("%w, getTotalStakedTopUpStakedBlsKeys function should have at least three values", epochStart.ErrExecutingSystemScCode)
	}

	topUpValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	totalStakedValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[1])

	return totalStakedValue, topUpValue, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakedValuesProc) IsInterfaceNil() bool {
	return svp == nil
}
