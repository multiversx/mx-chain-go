package intermediate

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgDeployProcessor is the argument used to create a deployLibrarySC instance
type ArgDeployLibrarySC struct {
	Executor         genesis.TxExecutionProcessor
	PubkeyConv       core.PubkeyConverter
	BlockchainHook   process.BlockChainHookHandler
	ShardCoordinator sharding.Coordinator
}

type deployLibrarySC struct {
	genesis.TxExecutionProcessor
	pubkeyConv       core.PubkeyConverter
	getScCodeAsHex   func(filename string) (string, error)
	blockchainHook   process.BlockChainHookHandler
	emptyAddress     []byte
	shardCoordinator sharding.Coordinator
}

// NewDeployLibrarySC returns a new instance of deploy library SC able to deploy library SC that needs to be present
// on all shards - same contract is deployed core.MaxShards time with addresses which end with all possibilities of the
// last 2 bytes
func NewDeployLibrarySC(arg ArgDeployLibrarySC) (*deployLibrarySC, error) {
	if check.IfNil(arg.Executor) {
		return nil, genesis.ErrNilTxExecutionProcessor
	}
	if check.IfNil(arg.PubkeyConv) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.BlockchainHook) {
		return nil, process.ErrNilBlockChainHook
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}

	dp := &deployLibrarySC{
		TxExecutionProcessor: arg.Executor,
		pubkeyConv:           arg.PubkeyConv,
		blockchainHook:       arg.BlockchainHook,
		shardCoordinator:     arg.ShardCoordinator,
	}
	dp.getScCodeAsHex = getSCCodeAsHex
	dp.emptyAddress = make([]byte, dp.pubkeyConv.Len())

	return dp, nil
}

// Deploy will try to deploy the provided smart contract
func (dp *deployLibrarySC) Deploy(sc genesis.InitialSmartContractHandler) error {
	code, err := dp.getScCodeAsHex(sc.GetFilename())
	if err != nil {
		return err
	}

	nonce, err := dp.GetNonce(sc.OwnerBytes())
	if err != nil {
		return err
	}

	scResultingAddressBytes, err := dp.blockchainHook.NewAddress(
		sc.OwnerBytes(),
		nonce,
		sc.VmTypeBytes(),
	)
	if err != nil {
		return err
	}

	sc.SetAddressBytes(scResultingAddressBytes)
	sc.SetAddress(dp.pubkeyConv.Encode(scResultingAddressBytes))

	vmType := sc.GetVmType()
	initParams := dp.applyCommonPlaceholders(sc.GetInitParameters())
	arguments := []string{code, vmType, codeMetadataHexForInitialSC}
	if len(initParams) > 0 {
		arguments = append(arguments, initParams)
	}
	deployTxData := strings.Join(arguments, "@")

	log.Trace("deploying genesis SC",
		"SC owner", sc.GetOwner(),
		"SC address", sc.Address(),
		"type", sc.GetType(),
		"VM type", sc.GetVmType(),
		"init params", initParams,
	)

	accountExists := dp.AccountExists(scResultingAddressBytes)
	if accountExists {
		return fmt.Errorf("%w for SC address %s, owner %s with nonce %d",
			genesis.ErrAccountAlreadyExists,
			sc.Address(),
			sc.GetOwner(),
			nonce,
		)
	}

	err = dp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		dp.emptyAddress,
		big.NewInt(0),
		[]byte(deployTxData),
	)
	if err != nil {
		return err
	}

	accountExists = dp.AccountExists(scResultingAddressBytes)
	if !accountExists {
		return fmt.Errorf("%w for SC address %s, owner %s with nonce %d",
			genesis.ErrAccountNotCreated,
			sc.Address(),
			sc.GetOwner(),
			nonce,
		)
	}

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (dp *deployLibrarySC) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
