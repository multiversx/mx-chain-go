package trieIterators

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type directStakedListProc struct {
	*commonStakingProcessor
	accounts           *AccountsWrapper
	blockChain         data.ChainHandler
	publicKeyConverter core.PubkeyConverter
}

var metachainIdentifier = []byte{255}

// NewDirectStakedListProcessor will create a new instance of stakedValuesProc
func NewDirectStakedListProcessor(
	accounts *AccountsWrapper,
	blockChain data.ChainHandler,
	queryService process.SCQueryService,
	publicKeyConverter core.PubkeyConverter,
) (*directStakedListProc, error) {
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
		return nil, fmt.Errorf("%w in NewDirectStakedListProcessor", ErrNilMutex)
	}

	return &directStakedListProc{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: queryService,
		},
		accounts:           accounts,
		blockChain:         blockChain,
		publicKeyConverter: publicKeyConverter,
	}, nil
}

// GetDirectStakedList will return the list for the direct staked addresses
func (dslp *directStakedListProc) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	dslp.accounts.Lock()
	defer dslp.accounts.Unlock()

	validatorAccount, err := dslp.getAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, err
	}

	return dslp.getAllStakedAccounts(validatorAccount)
}

func (dslp *directStakedListProc) getAllStakedAccounts(validatorAccount state.UserAccountHandler) ([]*api.DirectStakedValue, error) {
	rootHash, err := validatorAccount.DataTrie().RootHash()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	chLeaves, err := validatorAccount.DataTrie().GetAllLeavesOnChannel(rootHash, ctx)
	if err != nil {
		return nil, err
	}

	stakedAccounts := make([]*api.DirectStakedValue, 0)
	for leaf := range chLeaves {
		leafKey := leaf.Key()
		if len(leafKey) != dslp.publicKeyConverter.Len() {
			continue
		}
		if core.IsSmartContractOnMetachain(metachainIdentifier, leafKey) {
			continue
		}

		totalStakedCurrentAccount, totalTopUpCurrentAccount, errGet := dslp.getValidatorInfoFromSC(leafKey)
		if errGet != nil {
			continue
		}

		val := &api.DirectStakedValue{
			Address: dslp.publicKeyConverter.Encode(leafKey),
			Staked:  totalStakedCurrentAccount.String(),
			TopUp:   totalTopUpCurrentAccount.String(),
			Total:   big.NewInt(0).Add(totalTopUpCurrentAccount, totalStakedCurrentAccount).String(),
		}

		stakedAccounts = append(stakedAccounts, val)
	}

	return stakedAccounts, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dslp *directStakedListProc) IsInterfaceNil() bool {
	return dslp == nil
}
