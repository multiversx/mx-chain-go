package checking

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const minimumAcceptedNodePrice = 0

var zero = big.NewInt(0)

type nodeSetupChecker struct {
	accountsParser           genesis.AccountsParser
	initialNodePrice         *big.Int
	validatorPubkeyConverter core.PubkeyConverter
	keyGenerator             crypto.KeyGenerator
}

type delegationAddress struct {
	value   *big.Int
	address string
}

// NewNodesSetupChecker will create a node setup checker able to check the initial nodes against the provided genesis values
func NewNodesSetupChecker(
	accountsParser genesis.AccountsParser,
	initialNodePrice *big.Int,
	validatorPubkeyConverter core.PubkeyConverter,
	keyGenerator crypto.KeyGenerator,
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
	if check.IfNil(keyGenerator) {
		return nil, genesis.ErrNilKeyGenerator
	}

	return &nodeSetupChecker{
		accountsParser:           accountsParser,
		initialNodePrice:         initialNodePrice,
		validatorPubkeyConverter: validatorPubkeyConverter,
		keyGenerator:             keyGenerator,
	}, nil
}

// Check will check that each and every initial node has a backed staking address
// also, it checks that the amount staked (either directly or delegated) matches exactly the total
// staked value defined in the genesis file
func (nsc *nodeSetupChecker) Check(initialNodes []sharding.GenesisNodeInfoHandler) error {
	err := nsc.ckeckGenesisNodes(initialNodes)
	if err != nil {
		return err
	}

	initialAccounts := nsc.getClonedInitialAccounts()
	delegated := nsc.createDelegatedValues(initialAccounts)
	err = nsc.traverseInitialNodesSubtractingStakedValue(initialAccounts, initialNodes, delegated)
	if err != nil {
		return err
	}

	return nsc.checkRemainderInitialAccounts(initialAccounts, delegated)
}

func (nsc *nodeSetupChecker) ckeckGenesisNodes(initialNodes []sharding.GenesisNodeInfoHandler) error {
	for _, node := range initialNodes {
		err := nsc.keyGenerator.CheckPublicKeyValid(node.PubKeyBytes())
		if err != nil {
			return fmt.Errorf("%w for node's public key `%s`, error: %s",
				genesis.ErrInvalidPubKey,
				nsc.validatorPubkeyConverter.Encode(node.PubKeyBytes()),
				err.Error(),
			)
		}
	}

	return nil
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
	delegated map[string]*delegationAddress,
) error {
	for _, initialNode := range initialNodes {
		err := nsc.subtractStakedValue(initialNode.AddressBytes(), initialAccounts, delegated)
		if err != nil {
			return fmt.Errorf("'%w' while processing node pubkey %s",
				err, nsc.validatorPubkeyConverter.Encode(initialNode.PubKeyBytes()))
		}
	}

	return nil
}

func (nsc *nodeSetupChecker) subtractStakedValue(
	addressBytes []byte,
	initialAccounts []genesis.InitialAccountHandler,
	delegated map[string]*delegationAddress,
) error {

	for _, ia := range initialAccounts {
		if bytes.Equal(ia.AddressBytes(), addressBytes) {
			ia.GetStakingValue().Sub(ia.GetStakingValue(), nsc.initialNodePrice)
			if ia.GetStakingValue().Cmp(zero) < 0 {
				return genesis.ErrStakingValueIsNotEnough
			}

			return nil
		}

		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			return genesis.ErrNilDelegationHandler
		}
		if !bytes.Equal(dh.AddressBytes(), addressBytes) {
			continue
		}

		addr, ok := delegated[string(dh.AddressBytes())]
		if !ok {
			continue
		}

		addr.value.Sub(addr.value, nsc.initialNodePrice)
		if addr.value.Cmp(zero) < 0 {
			return genesis.ErrDelegationValueIsNotEnough
		}

		return nil
	}

	return genesis.ErrNodeNotStaked
}

// checkRemainderInitialAccounts checks that both staked value and delegated value is 0, meaning that all
// subtractions occurred perfectly
func (nsc *nodeSetupChecker) checkRemainderInitialAccounts(
	initialAccounts []genesis.InitialAccountHandler,
	delegated map[string]*delegationAddress,
) error {

	for _, ia := range initialAccounts {
		if ia.GetStakingValue().Cmp(zero) != 0 {
			return fmt.Errorf("%w for staking address %s, remainder %s",
				genesis.ErrInvalidStakingBalance, ia.GetAddress(), ia.GetStakingValue().String(),
			)
		}
	}

	for _, delegation := range delegated {
		if delegation.value.Cmp(zero) != 0 {
			return fmt.Errorf("%w for delegation address %s, remainder %s",
				genesis.ErrInvalidDelegationValue,
				delegation.address,
				delegation.value.String(),
			)
		}
	}

	return nil
}

func (nsc *nodeSetupChecker) createDelegatedValues(initialAccounts []genesis.InitialAccountHandler) map[string]*delegationAddress {
	delegated := make(map[string]*delegationAddress)

	for _, ia := range initialAccounts {
		delegation := ia.GetDelegationHandler()
		if check.IfNil(delegation) {
			continue
		}
		delegationAddressBytes := delegation.AddressBytes()
		if len(delegationAddressBytes) == 0 {
			continue
		}

		delegatedAddr := delegated[string(delegationAddressBytes)]
		if delegatedAddr == nil {
			delegatedAddr = &delegationAddress{
				address: delegation.GetAddress(),
				value:   big.NewInt(0),
			}

			delegated[string(delegationAddressBytes)] = delegatedAddr
		}

		delegatedAddr.value.Add(delegatedAddr.value, delegation.GetValue())
		delegation.GetValue().SetUint64(0)
	}

	return delegated
}

// IsInterfaceNil returns if underlying object is true
func (nsc *nodeSetupChecker) IsInterfaceNil() bool {
	return nsc == nil
}
