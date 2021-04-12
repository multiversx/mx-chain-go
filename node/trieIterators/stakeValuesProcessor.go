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

type stakedValuesProcessor struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
}

// ArgTrieIteratorProcessor represents the arguments DTO used in trie iterator processors constructors
type ArgTrieIteratorProcessor struct {
	ShardID            uint32
	Accounts           *AccountsWrapper
	BlockChain         data.ChainHandler
	QueryService       process.SCQueryService
	PublicKeyConverter core.PubkeyConverter
}

// NewTotalStakedValueProcessor will create a new instance of stakedValuesProc
func NewTotalStakedValueProcessor(arg ArgTrieIteratorProcessor) (*stakedValuesProcessor, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	return &stakedValuesProcessor{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: arg.QueryService,
			blockChain:   arg.BlockChain,
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
	}, nil
}

func checkArguments(arg ArgTrieIteratorProcessor) error {
	if arg.Accounts == nil || check.IfNil(arg.Accounts) {
		return ErrNilAccountsAdapter
	}
	if check.IfNil(arg.BlockChain) {
		return ErrNilBlockChain
	}
	if check.IfNil(arg.QueryService) {
		return ErrNilQueryService
	}
	if check.IfNil(arg.PublicKeyConverter) {
		return ErrNilPubkeyConverter
	}
	if arg.Accounts.Mutex == nil {
		return fmt.Errorf("%w in NewTotalStakedValueProcessor", ErrNilMutex)
	}

	return nil
}

// GetTotalStakedValue will calculate total staked value if needed and return calculated value
func (svp *stakedValuesProcessor) GetTotalStakedValue() (*api.StakeValues, error) {
	totalStaked, topUp, err := svp.computeStakedValueAndTopUp()
	if err != nil {
		return nil, err
	}

	return &api.StakeValues{
		TotalStaked: totalStaked,
		TopUp:       topUp,
	}, nil
}

func (svp *stakedValuesProcessor) computeStakedValueAndTopUp() (*big.Int, *big.Int, error) {
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
func (svp *stakedValuesProcessor) IsInterfaceNil() bool {
	return svp == nil
}
