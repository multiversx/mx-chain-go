package intermediate

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgDeployLibrarySC is the argument used to create a deployLibrarySC instance
type ArgDeployLibrarySC struct {
	Executor         genesis.TxExecutionProcessor
	PubkeyConv       core.PubkeyConverter
	BlockchainHook   process.BlockChainHookHandler
	ShardCoordinator sharding.Coordinator
}

type deployLibrarySC struct {
	*baseDeploy
	shardCoordinator sharding.Coordinator
}

// NewDeployLibrarySC returns a new instance of deploy library SC able to deploy library SC that needs to be present
// on all shards - same contract is deployed core.MaxNumShards == 256 times with addresses which end with all
// possibilities of the last byte
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

	base := &baseDeploy{
		TxExecutionProcessor: arg.Executor,
		pubkeyConv:           arg.PubkeyConv,
		blockchainHook:       arg.BlockchainHook,
		emptyAddress:         make([]byte, arg.PubkeyConv.Len()),
		getScCodeAsHex:       getSCCodeAsHex,
	}

	dp := &deployLibrarySC{
		baseDeploy:       base,
		shardCoordinator: arg.ShardCoordinator,
	}

	return dp, nil
}

// GenerateInitialPublicKeys will create public keys for given shard or for all shards in case of deployment to core.MaxNumShards
func GenerateInitialPublicKeys(
	baseAddress []byte,
	isForCurrentShard func(address []byte) bool,
) [][]byte {
	addressLen := len(baseAddress)

	newAddresses := make([][]byte, 0)
	i := uint8(0)
	for ; i < core.MaxNumShards-1; i++ {
		shardInBytes := []byte{0, i}
		tmpAddress := string(baseAddress[:(addressLen-core.ShardIdentiferLen)]) + string(shardInBytes)

		if isForCurrentShard([]byte(tmpAddress)) {
			newAddresses = append(newAddresses, []byte(tmpAddress))
		}
	}

	shardInBytes := []byte{0, i}
	tmpAddress := string(baseAddress[:(addressLen-core.ShardIdentiferLen)]) + string(shardInBytes)

	if isForCurrentShard([]byte(tmpAddress)) {
		newAddresses = append(newAddresses, []byte(tmpAddress))
	}

	return newAddresses
}

// Deploy will try to deploy the provided smart contract
func (dp *deployLibrarySC) Deploy(sc genesis.InitialSmartContractHandler) ([][]byte, error) {
	code, err := dp.getScCodeAsHex(sc.GetFilename())
	if err != nil {
		return nil, err
	}

	isForCurrentShard := func(address []byte) bool {
		return dp.shardCoordinator.SelfId() == dp.shardCoordinator.ComputeId(address)
	}

	resultingScAddresses := make([][]byte, 0)
	newAddresses := GenerateInitialPublicKeys(genesis.InitialDNSAddress, isForCurrentShard)
	for _, newAddress := range newAddresses {
		scAddress, errDeploy := dp.deployForOneAddress(sc, newAddress, code, sc.GetInitParameters())
		if errDeploy != nil {
			return nil, errDeploy
		}

		resultingScAddresses = append(resultingScAddresses, scAddress)
		err = dp.changeOwnerAddress(scAddress, newAddress, sc.OwnerBytes())
		if err != nil {
			return nil, err
		}
	}

	return resultingScAddresses, nil
}

func (dp *deployLibrarySC) changeOwnerAddress(
	scAddress []byte,
	currentOwner []byte,
	newOwner []byte,
) error {
	nonce, err := dp.GetNonce(currentOwner)
	if err != nil {
		return err
	}

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	err = dp.ExecuteTransaction(nonce, currentOwner, scAddress, big.NewInt(0), txData)
	if err != nil {
		return err
	}

	account, ok := dp.GetAccount(scAddress)
	if !ok {
		return genesis.ErrChangeOwnerAddressFailed
	}

	if !bytes.Equal(account.GetOwnerAddress(), newOwner) {
		return genesis.ErrChangeOwnerAddressFailed
	}

	return err
}

// IsInterfaceNil returns if underlying object is true
func (dp *deployLibrarySC) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
