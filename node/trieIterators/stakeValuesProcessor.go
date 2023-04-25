package trieIterators

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
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
func (svp *stakedValuesProcessor) GetTotalStakedValue(ctx context.Context) (*api.StakeValues, error) {
	baseStaked, topUp, err := svp.computeBaseStakedAndTopUp(ctx)
	if err != nil {
		return nil, err
	}

	return &api.StakeValues{
		BaseStaked: baseStaked,
		TopUp:      topUp,
	}, nil
}

func (svp *stakedValuesProcessor) computeBaseStakedAndTopUp(ctx context.Context) (*big.Int, *big.Int, error) {
	svp.accounts.Lock()
	defer svp.accounts.Unlock()

	validatorAccount, err := svp.getAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, nil, err
	}

	// TODO investigate if a call to GetAllLeavesKeysOnChannel (without values) might increase performance
	chLeaves := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = validatorAccount.GetAllLeaves(chLeaves, ctx)
	if err != nil {
		return nil, nil, err
	}

	totalBaseStaked, totalTopUp := big.NewInt(0), big.NewInt(0)
	for leaf := range chLeaves.LeavesChan {
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

	err = chLeaves.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, nil, err
	}

	if common.IsContextDone(ctx) {
		return nil, nil, ErrTrieOperationsTimeout
	}

	return totalBaseStaked, totalTopUp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakedValuesProcessor) IsInterfaceNil() bool {
	return svp == nil
}
