package checking

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/genesis"
	"math/big"
)

type nodesSetupCheckerFactory struct{}

func NewNodesSetupCheckerFactory() *nodesSetupCheckerFactory {
	return &nodesSetupCheckerFactory{}
}

// Create will create a node setup checker able to check the initial nodes against the provided genesis values
func (scf *nodesSetupCheckerFactory) Create(
	accountsParser genesis.AccountsParser,
	initialNodePrice *big.Int,
	validatorPubkeyConverter core.PubkeyConverter,
	keyGenerator crypto.KeyGenerator,
) (NodeSetupChecker, error) {
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

// IsInterfaceNil returns if underlying object is true
func (f *nodesSetupCheckerFactory) IsInterfaceNil() bool {
	return f == nil
}
