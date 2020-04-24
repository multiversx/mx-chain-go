package checking

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const minimumAcceptedNodePrice = 0

type nodeSetupChecker struct {
	genesisParser            genesis.GenesisParser
	initialNodePrice         *big.Int
	validatorPubkeyConverter state.PubkeyConverter
	zeroValue                *big.Int
}

// NewNodesSetupChecker will create a node setup checker able to check the initial nodes against the provided genesis values
func NewNodesSetupChecker(
	genesisParser genesis.GenesisParser,
	initialNodePrice *big.Int,
	validatorPubkeyConverter state.PubkeyConverter,
) (*nodeSetupChecker, error) {
	if check.IfNil(genesisParser) {
		return nil, genesis.ErrNilGenesisParser
	}
	if initialNodePrice == nil {
		return nil, genesis.ErrNilInitialNodePrice
	}
	if initialNodePrice.Cmp(big.NewInt(minimumAcceptedNodePrice)) < 0 {
		return nil, fmt.Errorf("%w, minimum accepted is %d",
			genesis.ErrInvalidInitialNodePrice, minimumAcceptedNodePrice)
	}
	if check.IfNil(validatorPubkeyConverter) {
		return nil, genesis.ErrNilPubkeyConverter
	}

	return &nodeSetupChecker{
		genesisParser:            genesisParser,
		initialNodePrice:         initialNodePrice,
		validatorPubkeyConverter: validatorPubkeyConverter,
		zeroValue:                big.NewInt(0),
	}, nil
}

// Check will check that each and every initial node has a backed staking address
// also, it checks that the amount staked (either directly or delegated) matches exactly the total
// staked value defined in the genesis file
func (nsc *nodeSetupChecker) Check(initialNodes []sharding.GenesisNodeInfoHandler) error {
	var err error

	initialAccounts := nsc.getClonedInitialAccounts()
	err = nsc.traverseInitialNodes(initialAccounts, initialNodes)
	if err != nil {
		return err
	}

	return nsc.checkRemainderInitialAccounts(initialAccounts)
}

func (nsc *nodeSetupChecker) getClonedInitialAccounts() []*genesis.InitialAccount {
	initialAccounts := nsc.genesisParser.InitialAccounts()
	clonedInitialAccounts := make([]*genesis.InitialAccount, len(initialAccounts))

	for idx, ia := range initialAccounts {
		clonedInitialAccounts[idx] = ia.Clone()
	}

	return clonedInitialAccounts
}

func (nsc *nodeSetupChecker) traverseInitialNodes(
	initialAccounts []*genesis.InitialAccount,
	initialNodes []sharding.GenesisNodeInfoHandler,
) error {
	for _, initialNode := range initialNodes {
		err := nsc.subtractValue(initialNode.AddressBytes(), initialAccounts)
		if err != nil {
			return fmt.Errorf("'%w' while processing node pubkey %s",
				err, nsc.validatorPubkeyConverter.Encode(initialNode.PubKeyBytes()))
		}
	}

	return nil
}

func (nsc *nodeSetupChecker) subtractValue(addressBytes []byte, initialAccounts []*genesis.InitialAccount) error {
	for _, ia := range initialAccounts {
		if bytes.Equal(ia.AddressBytes(), addressBytes) {
			ia.StakingValue.Sub(ia.StakingValue, nsc.initialNodePrice)
			if ia.StakingValue.Cmp(nsc.zeroValue) < 0 {
				return genesis.ErrStakingValueIsNotEnough
			}

			return nil
		}

		if bytes.Equal(ia.Delegation.AddressBytes(), addressBytes) {
			ia.Delegation.Value.Sub(ia.Delegation.Value, nsc.initialNodePrice)
			if ia.Delegation.Value.Cmp(nsc.zeroValue) < 0 {
				return genesis.ErrDelegationValueIsNotEnough
			}

			return nil
		}
	}

	return genesis.ErrNodeNotStaked
}

// checkRemainderInitialAccounts checks that both staked value and delegated value is 0, meaning that all
// subtractions occurred perfectly
func (nsc *nodeSetupChecker) checkRemainderInitialAccounts(initialAccounts []*genesis.InitialAccount) error {
	for _, ia := range initialAccounts {
		if ia.StakingValue.Cmp(nsc.zeroValue) != 0 {
			return fmt.Errorf("%w for staking address %s, remainder %s",
				genesis.ErrInvalidStakingBalance, ia.Address, ia.StakingValue.String(),
			)
		}

		if ia.Delegation.Value.Cmp(nsc.zeroValue) != 0 {
			return fmt.Errorf("%w for delegation address %s, remainder %s",
				genesis.ErrInvalidDelegationValue, ia.Address, ia.Delegation.Value.String(),
			)
		}
	}

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (nsc *nodeSetupChecker) IsInterfaceNil() bool {
	return nsc == nil
}
