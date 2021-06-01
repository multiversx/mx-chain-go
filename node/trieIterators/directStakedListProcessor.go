package trieIterators

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type directStakedListProcessor struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
}

var metachainIdentifier = []byte{255}

// NewDirectStakedListProcessor will create a new instance of stakedValuesProc
func NewDirectStakedListProcessor(arg ArgTrieIteratorProcessor) (*directStakedListProcessor, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	return &directStakedListProcessor{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: arg.QueryService,
			blockChain:   arg.BlockChain,
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
	}, nil
}

// GetDirectStakedList will return the list for the direct staked addresses
func (dslp *directStakedListProcessor) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	dslp.accounts.Lock()
	defer dslp.accounts.Unlock()

	validatorAccount, err := dslp.getAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, err
	}

	return dslp.getAllStakedAccounts(validatorAccount)
}

func (dslp *directStakedListProcessor) getAllStakedAccounts(validatorAccount state.UserAccountHandler) ([]*api.DirectStakedValue, error) {
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

		info, errGet := dslp.getValidatorInfoFromSC(leafKey)
		if errGet != nil {
			continue
		}

		baseStaked := big.NewInt(0).Set(info.totalStakedValue)
		baseStaked.Sub(baseStaked, info.topUpValue)
		val := &api.DirectStakedValue{
			Address:    dslp.publicKeyConverter.Encode(leafKey),
			BaseStaked: baseStaked.String(),
			TopUp:      info.topUpValue.String(),
			Total:      info.totalStakedValue.String(),
		}

		stakedAccounts = append(stakedAccounts, val)
	}

	return stakedAccounts, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dslp *directStakedListProcessor) IsInterfaceNil() bool {
	return dslp == nil
}
