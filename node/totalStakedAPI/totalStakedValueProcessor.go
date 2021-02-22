package totalStakedAPI

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type totalStakedValueProcessor struct {
	marshalizer      marshal.Marshalizer
	accounts         state.AccountsAdapter
	cacheDuration    time.Duration
	lastComputeTime  time.Time
	totalStakedValue *big.Int
	mutex            sync.RWMutex
}

// NewTotalStakedValueProcessor will create a new instance of totalStakedValueProcessor
func NewTotalStakedValueProcessor(
	marshalizer marshal.Marshalizer,
	cacheDuration time.Duration,
	accounts state.AccountsAdapter,
) (*totalStakedValueProcessor, error) {
	if cacheDuration <= 0 {
		return nil, ErrInvalidTotalStakedValueCacheDuration
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accounts) {
		return nil, ErrNilAccountsAdapter
	}

	return &totalStakedValueProcessor{
		marshalizer:      marshalizer,
		accounts:         accounts,
		cacheDuration:    cacheDuration,
		totalStakedValue: big.NewInt(0),
		mutex:            sync.RWMutex{},
		lastComputeTime:  time.Time{},
	}, nil
}

// GetTotalStakedValue will calculate total staked value if needed and return calculated value
func (tsp *totalStakedValueProcessor) GetTotalStakedValue() (*big.Int, error) {
	if time.Since(tsp.lastComputeTime) < tsp.cacheDuration {
		tsp.mutex.RLock()
		defer tsp.mutex.RUnlock()

		return tsp.totalStakedValue, nil
	}

	tsp.mutex.Lock()
	defer tsp.mutex.Unlock()
	err := tsp.updateTotalStakedValue()
	if err != nil {
		return nil, err
	}

	return tsp.totalStakedValue, nil
}

func (tsp *totalStakedValueProcessor) updateTotalStakedValue() error {
	ah, err := tsp.accounts.GetExistingAccount(vm.ValidatorSCAddress)
	if err != nil {
		return err
	}

	account, ok := ah.(state.UserAccountHandler)
	if !ok {
		return ErrCannotCastAccountHandlerToUserAccount
	}

	rootHash, err := account.DataTrie().RootHash()
	if err != nil {
		return err
	}

	ctx := context.Background()
	chLeaves, err := account.DataTrie().GetAllLeavesOnChannel(rootHash, ctx)
	if err != nil {
		return err
	}

	totalStakedValueAllLeaves := big.NewInt(0)
	for leaf := range chLeaves {
		validatorData := &systemSmartContracts.ValidatorDataV2{}
		value, errTrim := leaf.ValueWithoutSuffix(append(leaf.Key(), vm.ValidatorSCAddress...))
		if errTrim != nil {
			return fmt.Errorf("%w for validator key %s", errTrim, hex.EncodeToString(leaf.Key()))
		}

		err = tsp.marshalizer.Unmarshal(validatorData, value)
		if err != nil {
			continue
		}

		totalStakedValueAllLeaves.Add(totalStakedValueAllLeaves, validatorData.TotalStakeValue)
	}

	tsp.totalStakedValue = totalStakedValueAllLeaves
	tsp.lastComputeTime = time.Now()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsp *totalStakedValueProcessor) IsInterfaceNil() bool {
	return tsp == nil
}
