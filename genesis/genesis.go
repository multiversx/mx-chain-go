package genesis

import (
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
	initialAccounts []*InitialAccount
	entireSupply    *big.Int
	pubkeyConverter state.PubkeyConverter
}

// NewGenesis creates a new decoded genesis structure from json config file
func NewGenesis(
	genesisFilePath string,
	entireSupply *big.Int,
	pubkeyConverter state.PubkeyConverter,
) (*Genesis, error) {

	if entireSupply == nil {
		return nil, ErrNilEntireSupply
	}
	if big.NewInt(0).Cmp(entireSupply) >= 0 {
		return nil, ErrInvalidEntireSupply
	}
	if check.IfNil(pubkeyConverter) {
		return nil, ErrNilPubkeyConverter
	}

	initialAccounts := make([]*InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, genesisFilePath)
	if err != nil {
		return nil, err
	}

	genesis := &Genesis{
		initialAccounts: initialAccounts,
		entireSupply:    entireSupply,
		pubkeyConverter: pubkeyConverter,
	}

	err = genesis.process()
	if err != nil {
		return nil, err
	}

	return genesis, nil
}

func (g *Genesis) process() error {
	totalSupply := big.NewInt(0)
	for _, initialAccount := range g.initialAccounts {
		err := g.parseElement(initialAccount)
		if err != nil {
			return err
		}

		err = g.checkInitialAccount(initialAccount)
		if err != nil {
			return err
		}

		totalSupply.Add(totalSupply, initialAccount.Supply)
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

func (g *Genesis) parseElement(initialAccount *InitialAccount) error {
	var err error

	if len(initialAccount.Address) == 0 {
		return ErrEmptyAddress
	}
	initialAccount.address, err = g.pubkeyConverter.Decode(initialAccount.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`",
			ErrInvalidAddress, initialAccount.Address)
	}

	return g.parseDelegationElement(initialAccount)
}

func (g *Genesis) parseDelegationElement(initialAccount *InitialAccount) error {
	var err error
	delegationData := initialAccount.Delegation

	if big.NewInt(0).Cmp(delegationData.Value) == 0 {
		return nil
	}

	if len(delegationData.Address) == 0 {
		return fmt.Errorf("%w for address '%s'",
			ErrEmptyDelegationAddress, initialAccount.Address)
	}
	delegationData.address, err = g.pubkeyConverter.Decode(delegationData.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`, address %s",
			ErrInvalidDelegationAddress,
			delegationData.Address,
			initialAccount.Address,
		)
	}

	return nil
}

func (g *Genesis) checkInitialAccount(initialAccount *InitialAccount) error {
	isSmartContract := core.IsSmartContractAddress(initialAccount.address)
	if isSmartContract {
		return fmt.Errorf("%w for address %s",
			ErrAddressIsSmartContract,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Supply) >= 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidSupply,
			initialAccount.Supply,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Balance) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.StakingValue) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidStakingBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Delegation.Value) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidDelegationValue,
			initialAccount.Delegation.Value,
			initialAccount.Address,
		)
	}

	sum := big.NewInt(0)
	sum.Add(sum, initialAccount.Balance)
	sum.Add(sum, initialAccount.StakingValue)
	sum.Add(sum, initialAccount.Delegation.Value)

	isSupplyCorrect := big.NewInt(0).Cmp(initialAccount.Supply) < 0 && initialAccount.Supply.Cmp(sum) == 0
	if !isSupplyCorrect {
		return fmt.Errorf("%w for address %s, provided %s, computed %s",
			ErrSupplyMismatch,
			initialAccount.Address,
			initialAccount.Supply.String(),
			sum.String(),
		)
	}

	return nil
}

func (g *Genesis) checkForDuplicates() error {
	for idx1 := 0; idx1 < len(g.initialAccounts); idx1++ {
		ib1 := g.initialAccounts[idx1]
		for idx2 := idx1 + 1; idx2 < len(g.initialAccounts); idx2++ {
			ib2 := g.initialAccounts[idx2]
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
	for _, ib := range g.initialAccounts {
		if ib.Address == address {
			return big.NewInt(0).Set(ib.StakingValue)
		}
	}

	return big.NewInt(0)
}

// DelegatedUpon returns the value that was delegated upon the provided address
func (g *Genesis) DelegatedUpon(address string) *big.Int {
	delegated := big.NewInt(0)
	for _, ib := range g.initialAccounts {
		if ib.Delegation.Address == address {
			delegated.Add(delegated, ib.Delegation.Value)
		}
	}

	return delegated
}

// InitialAccountsSplitOnAddressesShards gets the initial accounts of the nodes split on the addresses's shards
func (g *Genesis) InitialAccountsSplitOnAddressesShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]*InitialAccount, error) {

	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	var addresses = make(map[uint32][]*InitialAccount)
	for _, in := range g.initialAccounts {
		address, err := g.pubkeyConverter.CreateAddressFromBytes(in.address)
		if err != nil {
			return nil, err
		}
		shardID := shardCoordinator.ComputeId(address)

		addresses[shardID] = append(addresses[shardID], in)
	}

	return addresses, nil
}

// InitialAccountsSplitOnDelegationAddressesShards gets the initial accounts of the nodes split on the addresses's shards
func (g *Genesis) InitialAccountsSplitOnDelegationAddressesShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]*InitialAccount, error) {

	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	var addresses = make(map[uint32][]*InitialAccount)
	for _, in := range g.initialAccounts {
		if len(in.Delegation.Address) == 0 {
			continue
		}

		delegationAddress, err := g.pubkeyConverter.CreateAddressFromBytes(in.Delegation.address)
		if err != nil {
			return nil, err
		}
		shardID := shardCoordinator.ComputeId(delegationAddress)

		addresses[shardID] = append(addresses[shardID], in)
	}

	return addresses, nil
}
