package genesis

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const decodeBase = 10

// Genesis hold data for decoded data from json file
type Genesis struct {
	initialBalances []*InitialBalance
	entireSupply    *big.Int
}

// NewGenesis creates a new decoded genesis structure from json config file
func NewGenesis(
	genesisFilePath string,
	entireSupply *big.Int,
) (*Genesis, error) {

	if entireSupply == nil {
		return nil, ErrNilEntireSupply
	}
	if big.NewInt(0).Cmp(entireSupply) >= 0 {
		return nil, ErrInvalidEntireSupply
	}

	initialBalances := make([]*InitialBalance, 0)
	err := core.LoadJsonFile(&initialBalances, genesisFilePath)
	if err != nil {
		return nil, err
	}

	genesis := &Genesis{
		initialBalances: initialBalances,
		entireSupply:    entireSupply,
	}

	err = genesis.process()
	if err != nil {
		return nil, err
	}

	return genesis, nil
}

func (g *Genesis) process() error {
	totalSupply := big.NewInt(0)
	for _, initialBalance := range g.initialBalances {
		err := g.tryParseElement(initialBalance)
		if err != nil {
			return err
		}

		err = g.checkInitialBalance(initialBalance)
		if err != nil {
			return err
		}

		totalSupply.Add(totalSupply, initialBalance.supply)
	}

	err := g.checkForDuplicates()
	if err != nil {
		return err
	}

	if totalSupply.Cmp(g.entireSupply) != 0 {
		return fmt.Errorf("%w for entire supply provided %s, computed %s",
			ErrEntireSupplyMismatch,
			g.entireSupply.String(),
			totalSupply.String(),
		)
	}

	return nil
}

func (g *Genesis) tryParseElement(initialBalance *InitialBalance) error {
	var ok bool
	var err error

	if len(initialBalance.Address) == 0 {
		return ErrEmptyAddress
	}
	//TODO change here (use Decode method) when merging with feat/bech32 branch
	initialBalance.address, err = hex.DecodeString(initialBalance.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`",
			ErrInvalidAddress, initialBalance.Address)
	}

	initialBalance.supply, ok = big.NewInt(0).SetString(initialBalance.Supply, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidSupplyString,
			initialBalance.Supply,
			initialBalance.Address,
		)
	}

	initialBalance.balance, ok = big.NewInt(0).SetString(initialBalance.Balance, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidBalanceString,
			initialBalance.Balance,
			initialBalance.Address,
		)
	}

	initialBalance.stakingBalance, ok = big.NewInt(0).SetString(initialBalance.StakingBalance, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidStakingBalanceString,
			initialBalance.StakingBalance,
			initialBalance.Address,
		)
	}

	return g.tryParseDelegationElement(initialBalance)
}

func (g *Genesis) tryParseDelegationElement(initialBalance *InitialBalance) error {
	var ok bool
	var err error
	delegationData := initialBalance.Delegation

	delegationData.value, ok = big.NewInt(0).SetString(delegationData.Value, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidDelegationValueString,
			delegationData.Value,
			delegationData.Address,
		)
	}

	if big.NewInt(0).Cmp(delegationData.value) == 0 {
		return nil
	}

	if len(delegationData.Address) == 0 {
		return fmt.Errorf("%w for address '%s'",
			ErrEmptyDelegationAddress, initialBalance.Address)
	}
	//TODO change here (use Decode method) when merging with feat/bech32 branch
	delegationData.address, err = hex.DecodeString(delegationData.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`, address %s",
			ErrInvalidDelegationAddress,
			delegationData.Address,
			initialBalance.Address,
		)
	}

	return nil
}

func (g *Genesis) checkInitialBalance(initialBalance *InitialBalance) error {
	isSmartContract := core.IsSmartContractAddress(initialBalance.address)
	if isSmartContract {
		return fmt.Errorf("%w for address %s",
			ErrAddressIsSmartContract,
			initialBalance.Address,
		)
	}

	if big.NewInt(0).Cmp(initialBalance.supply) >= 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidSupply,
			initialBalance.Supply,
			initialBalance.Address,
		)
	}

	if big.NewInt(0).Cmp(initialBalance.balance) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidBalance,
			initialBalance.Balance,
			initialBalance.Address,
		)
	}

	if big.NewInt(0).Cmp(initialBalance.stakingBalance) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidStakingBalance,
			initialBalance.Balance,
			initialBalance.Address,
		)
	}

	if big.NewInt(0).Cmp(initialBalance.Delegation.value) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidDelegationValue,
			initialBalance.Delegation.value,
			initialBalance.Address,
		)
	}

	sum := big.NewInt(0)
	sum.Add(sum, initialBalance.balance)
	sum.Add(sum, initialBalance.stakingBalance)
	sum.Add(sum, initialBalance.Delegation.value)

	isSupplyCorrect := big.NewInt(0).Cmp(initialBalance.supply) < 0 && initialBalance.supply.Cmp(sum) == 0
	if !isSupplyCorrect {
		return fmt.Errorf("%w for address %s, provided %s, computed %s",
			ErrSupplyMismatch,
			initialBalance.Address,
			initialBalance.supply.String(),
			sum.String(),
		)
	}

	return nil
}

func (g *Genesis) checkForDuplicates() error {
	for idx1 := 0; idx1 < len(g.initialBalances); idx1++ {
		ib1 := g.initialBalances[idx1]
		for idx2 := idx1 + 1; idx2 < len(g.initialBalances); idx2++ {
			ib2 := g.initialBalances[idx2]
			if ib1.Address == ib2.Address {
				return fmt.Errorf("%w found for '%s'",
					ErrDuplicateAddress,
					ib1.Address,
				)
			}
		}
	}

	return nil
}

// StakedUpon returns the value that was staked upon the provided address
func (g *Genesis) StakedUpon(address string) *big.Int {
	for _, ib := range g.initialBalances {
		if ib.Address == address {
			return big.NewInt(0).Set(ib.stakingBalance)
		}
	}

	return big.NewInt(0)
}

// DelegatedUpon returns the value that was delegated upon the provided address
func (g *Genesis) DelegatedUpon(address string) *big.Int {
	delegated := big.NewInt(0)
	for _, ib := range g.initialBalances {
		if ib.Delegation.Address == address {
			delegated.Add(delegated, ib.Delegation.value)
		}
	}

	return delegated
}

// InitialBalancesSplitOnAddresses gets the initial balances of the nodes split on the addresses's shards
func (g *Genesis) InitialBalancesSplitOnAddresses(
	shardCoordinator sharding.Coordinator,
	adrConv state.AddressConverter,
) (map[uint32][]*InitialBalance, error) {

	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(adrConv) {
		return nil, ErrNilAddressConverter
	}

	var balances = make(map[uint32][]*InitialBalance)
	for _, in := range g.initialBalances {
		address, err := adrConv.CreateAddressFromPublicKeyBytes(in.address)
		if err != nil {
			return nil, err
		}
		shardID := shardCoordinator.ComputeId(address)

		balances[shardID] = append(balances[shardID], in)
	}

	return balances, nil
}

// InitialBalancesSplitOnDelegationAddresses gets the initial balances of the nodes split on the addresses's shards
func (g *Genesis) InitialBalancesSplitOnDelegationAddresses(
	shardCoordinator sharding.Coordinator,
	adrConv state.AddressConverter,
) (map[uint32][]*InitialBalance, error) {

	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(adrConv) {
		return nil, ErrNilAddressConverter
	}

	var balances = make(map[uint32][]*InitialBalance)
	for _, in := range g.initialBalances {
		if len(in.Delegation.address) == 0 {
			continue
		}

		delegationAddress, err := adrConv.CreateAddressFromPublicKeyBytes(in.Delegation.address)
		if err != nil {
			return nil, err
		}
		shardID := shardCoordinator.ComputeId(delegationAddress)

		balances[shardID] = append(balances[shardID], in)
	}

	return balances, nil
}
