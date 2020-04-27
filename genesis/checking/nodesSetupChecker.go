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
	accountsParser           genesis.AccountsParser
	initialNodePrice         *big.Int
	validatorPubkeyConverter state.PubkeyConverter
	zeroValue                *big.Int
}

// NewNodesSetupChecker will create a node setup checker able to check the initial nodes against the provided genesis values
func NewNodesSetupChecker(
	accountsParser genesis.AccountsParser,
	initialNodePrice *big.Int,
	validatorPubkeyConverter state.PubkeyConverter,
) (*nodeSetupChecker, error) {
	if check.IfNil(accountsParser) {
		return nil, genesis.ErrNilAccountsParser
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
		accountsParser:           accountsParser,
		initialNodePrice:         initialNodePrice,
		validatorPubkeyConverter: validatorPubkeyConverter,
		zeroValue:                big.NewInt(0),
	}, nil
}

// Check will check that each and every initial node has a backed staking address
// also, it checks that the amount staked (either directly or delegated) matches exactly the total
// staked value defined in the genesis file
func (nsc *nodeSetupChecker) Check(initialNodes []sharding.GenesisNodeInfoHandler) error {
	initialAccounts := nsc.getClonedInitialAccounts()
	err := nsc.traverseInitialNodesSubtractingStakedValue(initialAccounts, initialNodes)
	if err != nil {
		return err
	}

	return nsc.checkRemainderInitialAccounts(initialAccounts)
}

func (nsc *nodeSetupChecker) getClonedInitialAccounts() []genesis.InitialAccountHandler {
	initialAccounts := nsc.accountsParser.InitialAccounts()
	clonedInitialAccounts := make([]genesis.InitialAccountHandler, len(initialAccounts))

	for idx, ia := range initialAccounts {
		clonedInitialAccounts[idx] = ia.Clone()
	}

	return clonedInitialAccounts
}

func (nsc *nodeSetupChecker) traverseInitialNodesSubtractingStakedValue(
	initialAccounts []genesis.InitialAccountHandler,
	initialNodes []sharding.GenesisNodeInfoHandler,
) error {
	for _, initialNode := range initialNodes {
		err := nsc.subtractStakedValue(initialNode.AddressBytes(), initialAccounts)
		if err != nil {
			return fmt.Errorf("'%w' while processing node pubkey %s",
				err, nsc.validatorPubkeyConverter.Encode(initialNode.PubKeyBytes()))
		}
	}

	return nil
}

func (nsc *nodeSetupChecker) subtractStakedValue(addressBytes []byte, initialAccounts []genesis.InitialAccountHandler) error {
	for _, ia := range initialAccounts {
		if bytes.Equal(ia.AddressBytes(), addressBytes) {
			ia.GetStakingValue().Sub(ia.GetStakingValue(), nsc.initialNodePrice)
			if ia.GetStakingValue().Cmp(nsc.zeroValue) < 0 {
				return genesis.ErrStakingValueIsNotEnough
			}

			return nil
		}

		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			return genesis.ErrNilDelegationHandler
		}

		if bytes.Equal(dh.AddressBytes(), addressBytes) {
			dh.GetValue().Sub(dh.GetValue(), nsc.initialNodePrice)
			if dh.GetValue().Cmp(nsc.zeroValue) < 0 {
				return genesis.ErrDelegationValueIsNotEnough
			}

			return nil
		}
	}

	return genesis.ErrNodeNotStaked
}

// checkRemainderInitialAccounts checks that both staked value and delegated value is 0, meaning that all
// subtractions occurred perfectly
func (nsc *nodeSetupChecker) checkRemainderInitialAccounts(initialAccounts []genesis.InitialAccountHandler) error {
	for _, ia := range initialAccounts {
		if ia.GetStakingValue().Cmp(nsc.zeroValue) != 0 {
			return fmt.Errorf("%w for staking address %s, remainder %s",
				genesis.ErrInvalidStakingBalance, ia.GetAddress(), ia.GetStakingValue().String(),
			)
		}

		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			return genesis.ErrNilDelegationHandler
		}

		if dh.GetValue().Cmp(nsc.zeroValue) != 0 {
			return fmt.Errorf("%w for delegation address %s, remainder %s",
				genesis.ErrInvalidDelegationValue, ia.GetAddress(), dh.GetValue().String(),
			)
		}
	}

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (nsc *nodeSetupChecker) IsInterfaceNil() bool {
	return nsc == nil
}
