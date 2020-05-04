package parsing

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// accountsParser hold data for initial accounts decoded data from json file
type accountsParser struct {
	initialAccounts []*data.InitialAccount
	entireSupply    *big.Int
	pubkeyConverter state.PubkeyConverter
}

// NewAccountsParser creates a new decoded accounts genesis structure from json config file
func NewAccountsParser(
	genesisFilePath string,
	entireSupply *big.Int,
	pubkeyConverter state.PubkeyConverter,
) (*accountsParser, error) {

	if entireSupply == nil {
		return nil, genesis.ErrNilEntireSupply
	}
	if big.NewInt(0).Cmp(entireSupply) >= 0 {
		return nil, genesis.ErrInvalidEntireSupply
	}
	if check.IfNil(pubkeyConverter) {
		return nil, genesis.ErrNilPubkeyConverter
	}

	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, genesisFilePath)
	if err != nil {
		return nil, err
	}

	gp := &accountsParser{
		initialAccounts: initialAccounts,
		entireSupply:    entireSupply,
		pubkeyConverter: pubkeyConverter,
	}

	err = gp.process()
	if err != nil {
		return nil, err
	}

	return gp, nil
}

func (ap *accountsParser) process() error {
	totalSupply := big.NewInt(0)
	for _, initialAccount := range ap.initialAccounts {
		err := ap.parseElement(initialAccount)
		if err != nil {
			return err
		}

		err = ap.checkInitialAccount(initialAccount)
		if err != nil {
			return err
		}

		totalSupply.Add(totalSupply, initialAccount.Supply)
	}

	err := ap.checkForDuplicates()
	if err != nil {
		return err
	}

	if totalSupply.Cmp(ap.entireSupply) != 0 {
		return fmt.Errorf("%w for entire supply provided %s, computed %s",
			genesis.ErrEntireSupplyMismatch,
			ap.entireSupply.String(),
			totalSupply.String(),
		)
	}

	return nil
}

func (ap *accountsParser) parseElement(initialAccount *data.InitialAccount) error {
	if len(initialAccount.Address) == 0 {
		return genesis.ErrEmptyAddress
	}
	addressBytes, err := ap.pubkeyConverter.Decode(initialAccount.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`",
			genesis.ErrInvalidAddress, initialAccount.Address)
	}

	initialAccount.SetAddressBytes(addressBytes)

	return ap.parseDelegationElement(initialAccount)
}

func (ap *accountsParser) parseDelegationElement(initialAccount *data.InitialAccount) error {
	delegationData := initialAccount.Delegation

	if big.NewInt(0).Cmp(delegationData.Value) == 0 {
		return nil
	}

	if len(delegationData.Address) == 0 {
		return fmt.Errorf("%w for address '%s'",
			genesis.ErrEmptyDelegationAddress, initialAccount.Address)
	}
	addressBytes, err := ap.pubkeyConverter.Decode(delegationData.Address)
	if err != nil {
		return fmt.Errorf("%w for `%s`, address %s",
			genesis.ErrInvalidDelegationAddress,
			delegationData.Address,
			initialAccount.Address,
		)
	}

	delegationData.SetAddressBytes(addressBytes)

	return nil
}

func (ap *accountsParser) checkInitialAccount(initialAccount *data.InitialAccount) error {
	isSmartContract := core.IsSmartContractAddress(initialAccount.AddressBytes())
	if isSmartContract {
		return fmt.Errorf("%w for address %s",
			genesis.ErrAddressIsSmartContract,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Supply) >= 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidSupply,
			initialAccount.Supply,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Balance) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.StakingValue) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidStakingBalance,
			initialAccount.Balance,
			initialAccount.Address,
		)
	}

	if big.NewInt(0).Cmp(initialAccount.Delegation.Value) > 0 {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidDelegationValue,
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
			genesis.ErrSupplyMismatch,
			initialAccount.Address,
			initialAccount.Supply.String(),
			sum.String(),
		)
	}

	return nil
}

func (ap *accountsParser) checkForDuplicates() error {
	for idx1 := 0; idx1 < len(ap.initialAccounts); idx1++ {
		ia1 := ap.initialAccounts[idx1]
		for idx2 := idx1 + 1; idx2 < len(ap.initialAccounts); idx2++ {
			ia2 := ap.initialAccounts[idx2]
			if ia1.Address == ia2.Address {
				return fmt.Errorf("%w found for '%s'",
					genesis.ErrDuplicateAddress,
					ia1.Address,
				)
			}
		}
	}

	return nil
}

// StakedUpon returns the value that was staked upon the provided address
func (ap *accountsParser) StakedUpon(address string) *big.Int {
	for _, ia := range ap.initialAccounts {
		if ia.Address == address {
			return big.NewInt(0).Set(ia.StakingValue)
		}
	}

	return big.NewInt(0)
}

// DelegatedUpon returns the value that was delegated upon the provided address
func (ap *accountsParser) DelegatedUpon(address string) *big.Int {
	delegated := big.NewInt(0)
	for _, ia := range ap.initialAccounts {
		if ia.Delegation.Address == address {
			delegated.Add(delegated, ia.Delegation.Value)
		}
	}

	return delegated
}

// InitialAccounts return the initial accounts contained by this parser
func (ap *accountsParser) InitialAccounts() []genesis.InitialAccountHandler {
	accounts := make([]genesis.InitialAccountHandler, len(ap.initialAccounts))

	for idx, ia := range ap.initialAccounts {
		accounts[idx] = ia
	}

	return accounts
}

// InitialAccountsSplitOnAddressesShards gets the initial accounts of the nodes split on the addresses's shards
func (ap *accountsParser) InitialAccountsSplitOnAddressesShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]genesis.InitialAccountHandler, error) {

	if check.IfNil(shardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}

	var addresses = make(map[uint32][]genesis.InitialAccountHandler)
	for _, ia := range ap.initialAccounts {
		shardID := shardCoordinator.ComputeId(ia.AddressBytes())

		addresses[shardID] = append(addresses[shardID], ia)
	}

	return addresses, nil
}

// InitialAccountsSplitOnDelegationAddressesShards gets the initial accounts of the nodes split on the addresses's shards
func (ap *accountsParser) InitialAccountsSplitOnDelegationAddressesShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]genesis.InitialAccountHandler, error) {

	if check.IfNil(shardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}

	var addresses = make(map[uint32][]genesis.InitialAccountHandler)
	for _, in := range ap.initialAccounts {
		if len(in.Delegation.Address) == 0 {
			continue
		}

		shardID := shardCoordinator.ComputeId(in.Delegation.AddressBytes())

		addresses[shardID] = append(addresses[shardID], in)
	}

	return addresses, nil
}

// GetTotalStakedForDelegationAddress returns the total staked value for a provided delegation address
func (ap *accountsParser) GetTotalStakedForDelegationAddress(delegationAddress string) *big.Int {
	sum := big.NewInt(0)

	for _, in := range ap.initialAccounts {
		if in.Delegation.Address == delegationAddress {
			sum.Add(sum, in.Delegation.Value)
		}
	}

	return sum
}

// GetInitialAccountsForDelegated returns the initial accounts that are set to the provided delegated address
func (ap *accountsParser) GetInitialAccountsForDelegated(addressBytes []byte) []genesis.InitialAccountHandler {
	//TODO add implementation here
	return make([]genesis.InitialAccountHandler, 0)
}

// IsInterfaceNil returns if underlying object is true
func (ap *accountsParser) IsInterfaceNil() bool {
	return ap == nil
}
