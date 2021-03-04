package stakeValuesProcessor

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type stakedValuesProc struct {
	marshalizer marshal.Marshalizer
	accounts    state.AccountsAdapter
	nodePrice   *big.Int
}

// NewTotalStakedValueProcessor will create a new instance of stakedValuesProc
func NewTotalStakedValueProcessor(
	nodePrice string,
	marshalizer marshal.Marshalizer,
	accounts state.AccountsAdapter,
) (*stakedValuesProc, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accounts) {
		return nil, ErrNilAccountsAdapter
	}

	nodePriceBig, ok := big.NewInt(0).SetString(nodePrice, 10)
	if !ok {
		return nil, ErrInvalidNodePrice
	}

	return &stakedValuesProc{
		marshalizer: marshalizer,
		accounts:    accounts,
		nodePrice:   nodePriceBig,
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

	numRegistedNodes := uint64(0)
	totalStakedValueAllLeaves := big.NewInt(0)
	for leaf := range chLeaves {
		validatorData := &systemSmartContracts.ValidatorDataV2{}
		value, errTrim := leaf.ValueWithoutSuffix(append(leaf.Key(), vm.ValidatorSCAddress...))
		if errTrim != nil {
			return nil, nil, fmt.Errorf("%w for validator key %s", errTrim, hex.EncodeToString(leaf.Key()))
		}

		err = svp.marshalizer.Unmarshal(validatorData, value)
		if err != nil {
			continue
		}

		totalStakedValueAllLeaves.Add(totalStakedValueAllLeaves, validatorData.TotalStakeValue)
		numRegistedNodes += uint64(validatorData.NumRegistered)
	}

	totalStakedValue := totalStakedValueAllLeaves

	numRegisteredNodesBig := big.NewInt(0).SetUint64(numRegistedNodes)
	totalStakedValueWithoutTopUp := big.NewInt(0).Mul(numRegisteredNodesBig, svp.nodePrice)
	topUpValue := big.NewInt(0).Sub(totalStakedValueAllLeaves, totalStakedValueWithoutTopUp)

	return totalStakedValue, topUpValue, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakedValuesProc) IsInterfaceNil() bool {
	return svp == nil
}
