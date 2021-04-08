package trieIterators

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// AccountsWrapper extends the AccountsAdapter interface
type AccountsWrapper struct {
	*sync.Mutex
	state.AccountsAdapter
}

type stakedValuesProc struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
}

// NewTotalStakedValueProcessor will create a new instance of stakedValuesProc
func NewTotalStakedValueProcessor(
	accounts *AccountsWrapper,
	blockChain data.ChainHandler,
	queryService process.SCQueryService,
	publicKeyConverter core.PubkeyConverter,
) (*stakedValuesProc, error) {
	if accounts == nil || check.IfNil(accounts) {
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
	if accounts.Mutex == nil {
		return nil, fmt.Errorf("%w in NewTotalStakedValueProcessor", ErrNilMutex)
	}

	return &stakedValuesProc{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: queryService,
			blockChain:   blockChain,
			accounts:     accounts,
		},
		publicKeyConverter: publicKeyConverter,
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
	svp.accounts.Lock()
	defer svp.accounts.Unlock()

	validatorAccount, err := svp.getAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, nil, err
	}

	rootHash, err := validatorAccount.DataTrie().RootHash()
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	//TODO investigate if a call to GetAllLeavesKeysOnChannel (without values) might increase performance
	chLeaves, err := validatorAccount.DataTrie().GetAllLeavesOnChannel(rootHash, ctx)
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

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakedValuesProc) IsInterfaceNil() bool {
	return svp == nil
}
