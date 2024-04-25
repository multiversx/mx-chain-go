package checking

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

type NodeSetupCheckerFactory interface {
	Create(
		accountsParser genesis.AccountsParser,
		initialNodePrice *big.Int,
		validatorPubkeyConverter core.PubkeyConverter,
		keyGenerator crypto.KeyGenerator,
	) (NodeSetupChecker, error)
	IsInterfaceNil() bool
}

type NodeSetupChecker interface {
	Check(initialNodes []nodesCoordinator.GenesisNodeInfoHandler) error
	IsInterfaceNil() bool
}
