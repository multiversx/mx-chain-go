package trieIterators

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
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
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
	}, nil
}

func checkArguments(arg ArgTrieIteratorProcessor) error {
	if arg.Accounts == nil || check.IfNil(arg.Accounts) {
		return ErrNilAccountsAdapter
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
	baseStaked, topUp, err := svp.computeBaseStakedAndTopUp()
	if err != nil {
		return nil, err
	}

	return &api.StakeValues{
		BaseStaked: baseStaked,
		TopUp:      topUp,
	}, nil
}

func (svp *stakedValuesProcessor) computeBaseStakedAndTopUp() (*big.Int, *big.Int, error) {
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

	// TODO investigate if a call to GetAllLeavesKeysOnChannel (without values) might increase performance
	chLeaves, err := validatorAccount.DataTrie().GetAllLeavesOnChannel(rootHash)
	if err != nil {
		return nil, nil, err
	}

	totalBaseStaked, totalTopUp := big.NewInt(0), big.NewInt(0)
	for leaf := range chLeaves {
		leafKey := leaf.Key()
		if len(leafKey) != svp.publicKeyConverter.Len() {
			continue
		}

		info, errGet := svp.getValidatorInfoFromSC(leafKey)
		if errGet != nil {
			continue
		}
		baseStaked := big.NewInt(0).Set(info.totalStakedValue)
		baseStaked.Sub(baseStaked, info.topUpValue)

		totalBaseStaked = totalBaseStaked.Add(totalBaseStaked, baseStaked)
		totalTopUp = totalTopUp.Add(totalTopUp, info.topUpValue)
	}

	return totalBaseStaked, totalTopUp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakedValuesProcessor) IsInterfaceNil() bool {
	return svp == nil
}
