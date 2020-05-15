package intermediate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgDelegationProcessor is the argument used to construct a delegation processor
type ArgDelegationProcessor struct {
	Executor            genesis.TxExecutionProcessor
	ShardCoordinator    sharding.Coordinator
	AccountsParser      genesis.AccountsParser
	SmartContractParser genesis.InitialSmartContractParser
	NodesListSplitter   genesis.NodesListSplitter
	QueryService        external.SCQueryService
	NodePrice           *big.Int
}

const stakeFunction = "stakeGenesis"
const setBlsKeysFunction = "setBlsKeys"
const activateBlsKeysFunction = "activate"
const setNumNodesFunction = "setNumNodes"
const setStakePerNodeFunction = "setStakePerNode"

var log = logger.GetOrCreate("genesis/process/intermediate")
var zero = big.NewInt(0)

type delegationProcessor struct {
	genesis.TxExecutionProcessor
	shardCoordinator     sharding.Coordinator
	accuntsParser        genesis.AccountsParser
	smartContractsParser genesis.InitialSmartContractParser
	nodesListSplitter    genesis.NodesListSplitter
	queryService         external.SCQueryService
	nodePrice            *big.Int
}

// NewDelegationProcessor returns a new delegation processor instance
func NewDelegationProcessor(arg ArgDelegationProcessor) (*delegationProcessor, error) {
	if check.IfNil(arg.Executor) {
		return nil, genesis.ErrNilTxExecutionProcessor
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}
	if check.IfNil(arg.AccountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}
	if check.IfNil(arg.SmartContractParser) {
		return nil, genesis.ErrNilSmartContractParser
	}
	if check.IfNil(arg.NodesListSplitter) {
		return nil, genesis.ErrNilNodesListSplitter
	}
	if check.IfNil(arg.QueryService) {
		return nil, genesis.ErrNilQueryService
	}
	if arg.NodePrice == nil {
		return nil, genesis.ErrNilInitialNodePrice
	}
	if arg.NodePrice.Cmp(zero) <= 0 {
		return nil, genesis.ErrInvalidInitialNodePrice
	}

	return &delegationProcessor{
		TxExecutionProcessor: arg.Executor,
		shardCoordinator:     arg.ShardCoordinator,
		accuntsParser:        arg.AccountsParser,
		smartContractsParser: arg.SmartContractParser,
		nodesListSplitter:    arg.NodesListSplitter,
		queryService:         arg.QueryService,
		nodePrice:            arg.NodePrice,
	}, nil
}

// ExecuteDelegation will execute stake, set bls keys and activate on all delegation contracts from this shard
func (dp *delegationProcessor) ExecuteDelegation() (genesis.DelegationResult, error) {
	smartContracts, err := dp.getDelegationScOnCurrentShard()
	if err != nil {
		return genesis.DelegationResult{}, err
	}
	if len(smartContracts) == 0 {
		return genesis.DelegationResult{}, nil
	}

	err = dp.setDelegationStartParameters(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	_, err = dp.executeManageBlsKeys(smartContracts, dp.getBlsKey, setBlsKeysFunction)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	dr := genesis.DelegationResult{}
	dr.NumTotalStaked, err = dp.executeStake(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	dr.NumTotalDelegated, err = dp.executeManageBlsKeys(smartContracts, dp.getBlsKeySig, activateBlsKeysFunction)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	err = dp.executeVerify(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	return dr, err
}

func (dp *delegationProcessor) getDelegationScOnCurrentShard() ([]genesis.InitialSmartContractHandler, error) {
	allSmartContracts, err := dp.smartContractsParser.InitialSmartContractsSplitOnOwnersShards(dp.shardCoordinator)
	if err != nil {
		return nil, err
	}

	smartContracts := make([]genesis.InitialSmartContractHandler, 0)
	smartContractsForCurrentShard := allSmartContracts[dp.shardCoordinator.SelfId()]
	for _, sc := range smartContractsForCurrentShard {
		if sc.GetType() == genesis.DelegationType {
			smartContracts = append(smartContracts, sc)
		}
	}

	log.Trace("getDelegationScOnCurrentShard",
		"num delegation SC", len(smartContracts),
		"shard ID", dp.shardCoordinator.SelfId(),
	)
	return smartContracts, nil
}

func (dp *delegationProcessor) setDelegationStartParameters(smartContracts []genesis.InitialSmartContractHandler) error {
	for _, sc := range smartContracts {
		delegatedNodes := dp.nodesListSplitter.GetDelegatedNodes(sc.AddressBytes())
		numNodes := len(delegatedNodes)

		log.Trace("setDelegationStartParameters",
			"SC owner", sc.GetOwner(),
			"SC address", sc.Address(),
			"num delegated nodes", numNodes,
			"node price", dp.nodePrice.String(),
			"shard ID", dp.shardCoordinator.SelfId(),
		)

		err := dp.executeSetNumNodes(numNodes, sc)
		if err != nil {
			return err
		}

		err = dp.executeSetNodePrice(sc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dp *delegationProcessor) executeSetNumNodes(numNodes int, sc genesis.InitialSmartContractHandler) error {
	setNumNodesTxData := fmt.Sprintf("%s@%x", setNumNodesFunction, numNodes)

	nonce, err := dp.GetNonce(sc.OwnerBytes())
	if err != nil {
		return err
	}

	return dp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		sc.AddressBytes(),
		zero,
		[]byte(setNumNodesTxData),
	)
}

func (dp *delegationProcessor) executeSetNodePrice(sc genesis.InitialSmartContractHandler) error {
	setStakePerNodeTxData := fmt.Sprintf("%s@%x", setStakePerNodeFunction, dp.nodePrice)

	nonce, err := dp.GetNonce(sc.OwnerBytes())
	if err != nil {
		return err
	}

	return dp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		sc.AddressBytes(),
		zero,
		[]byte(setStakePerNodeTxData),
	)
}

func (dp *delegationProcessor) executeStake(smartContracts []genesis.InitialSmartContractHandler) (int, error) {
	stakedOnDelegation := 0

	for _, sc := range smartContracts {
		accounts := dp.accuntsParser.GetInitialAccountsForDelegated(sc.AddressBytes())
		if len(accounts) == 0 {
			log.Debug("genesis delegation SC was not delegated by any account",
				"SC owner", sc.GetOwner(),
				"SC address", sc.Address(),
			)
			continue
		}

		totalDelegated := big.NewInt(0)
		for _, ac := range accounts {
			err := dp.stake(ac, sc)
			if err != nil {
				return 0, fmt.Errorf("%w while calling stake function from account %s", err, ac.GetAddress())
			}

			totalDelegated.Add(totalDelegated, ac.GetDelegationHandler().GetValue())
		}

		log.Trace("executeStake",
			"SC owner", sc.GetOwner(),
			"SC address", sc.Address(),
			"num accounts", len(accounts),
			"total delegated", totalDelegated,
		)
		stakedOnDelegation += len(accounts)
	}

	return stakedOnDelegation, nil
}

func (dp *delegationProcessor) stake(ac genesis.InitialAccountHandler, sc genesis.InitialSmartContractHandler) error {
	isIntraShardCall := dp.shardCoordinator.SameShard(ac.AddressBytes(), sc.AddressBytes())

	dh := ac.GetDelegationHandler()
	if check.IfNil(dh) {
		return genesis.ErrNilDelegationHandler
	}
	if dh.GetValue() == nil {
		return genesis.ErrInvalidDelegationValue
	}

	var err error
	var nonce = uint64(0)
	if isIntraShardCall {
		//intra shard transaction, get current nonce in order to make the tx processor work
		nonce, err = dp.GetNonce(ac.AddressBytes())
		if err != nil {
			return err
		}
	}

	stakeData := fmt.Sprintf("%s@%s", stakeFunction, dh.GetValue().Text(16))
	err = dp.ExecuteTransaction(
		nonce,
		ac.AddressBytes(),
		sc.AddressBytes(),
		zero,
		[]byte(stakeData),
	)
	if err != nil {
		return err
	}

	return nil
}

func (dp *delegationProcessor) executeManageBlsKeys(
	smartContracts []genesis.InitialSmartContractHandler,
	handler func(node sharding.GenesisNodeInfoHandler) string,
	function string,
) (int, error) {

	log.Trace("executeManageSetBlsKeys",
		"num delegation SC", len(smartContracts),
		"shard ID", dp.shardCoordinator.SelfId(),
		"function", function,
	)

	totalDelegated := 0
	for _, sc := range smartContracts {
		delegatedNodes := dp.nodesListSplitter.GetDelegatedNodes(sc.AddressBytes())

		lenDelegated := len(delegatedNodes)
		if lenDelegated == 0 {
			log.Debug("genesis delegation SC does not have staked nodes",
				"SC owner", sc.GetOwner(),
				"SC address", sc.Address(),
				"function", function,
			)
			continue
		}
		totalDelegated += lenDelegated

		log.Trace("executeSetBlsKeys",
			"SC owner", sc.GetOwner(),
			"SC address", sc.Address(),
			"num nodes", lenDelegated,
			"shard ID", dp.shardCoordinator.SelfId(),
			"function", function,
		)

		arguments := make([]string, 0, len(delegatedNodes)+1)
		arguments = append(arguments, function)
		for _, node := range delegatedNodes {
			arg := handler(node)
			arguments = append(arguments, arg)
		}

		nonce, err := dp.GetNonce(sc.OwnerBytes())
		if err != nil {
			return 0, err
		}

		err = dp.ExecuteTransaction(
			nonce,
			sc.OwnerBytes(),
			sc.AddressBytes(),
			big.NewInt(0),
			[]byte(strings.Join(arguments, "@")),
		)
		if err != nil {
			return 0, err
		}
	}

	return totalDelegated, nil
}

func (dp *delegationProcessor) getBlsKey(node sharding.GenesisNodeInfoHandler) string {
	return hex.EncodeToString(node.PubKeyBytes())
}

func (dp *delegationProcessor) getBlsKeySig(_ sharding.GenesisNodeInfoHandler) string {
	mockSignature := []byte("genesis signature")

	return hex.EncodeToString(mockSignature)
}

func (dp *delegationProcessor) executeVerify(smartContracts []genesis.InitialSmartContractHandler) error {
	for _, sc := range smartContracts {
		err := dp.verify(sc)
		if err != nil {
			return fmt.Errorf("%w for contract %s, owner %s", err, sc.Address(), sc.GetOwner())
		}
	}

	return nil
}

func (dp *delegationProcessor) verify(sc genesis.InitialSmartContractHandler) error {
	err := dp.verifyStakedValue(sc)
	if err != nil {
		return fmt.Errorf("%w for verifyStakedValue", err)
	}

	err = dp.verifyRegisteredNodes(sc)
	if err != nil {
		return fmt.Errorf("%w for verifyRegisteredNodes", err)
	}

	return nil
}

func (dp *delegationProcessor) verifyStakedValue(sc genesis.InitialSmartContractHandler) error {
	scQueryStakeValue := &process.SCQuery{
		ScAddress: sc.AddressBytes(),
		FuncName:  "getFilledStake",
		Arguments: [][]byte{},
	}
	vmOutputStakeValue, err := dp.queryService.ExecuteQuery(scQueryStakeValue)
	if err != nil {
		return err
	}
	if len(vmOutputStakeValue.ReturnData) != 1 {
		return fmt.Errorf("%w return data should have contained one element", genesis.ErrWhileVerifyingDelegation)
	}
	scStakedValue := big.NewInt(0).SetBytes(vmOutputStakeValue.ReturnData[0])
	providedStakedValue := big.NewInt(0)
	providedDelegators := dp.accuntsParser.GetInitialAccountsForDelegated(sc.AddressBytes())

	for _, delegator := range providedDelegators {
		if check.IfNil(delegator) {
			continue
		}
		dh := delegator.GetDelegationHandler()
		if check.IfNil(dh) {
			continue
		}
		if dh.GetValue() == nil {
			continue
		}
		providedStakedValue.Add(providedStakedValue, dh.GetValue())
	}
	if scStakedValue.Cmp(providedStakedValue) != 0 {
		return fmt.Errorf("%w staked data mismatch: from SC: %s, provided: %s",
			genesis.ErrWhileVerifyingDelegation, scStakedValue.String(), providedStakedValue.String())
	}

	return nil
}

func (dp *delegationProcessor) verifyRegisteredNodes(sc genesis.InitialSmartContractHandler) error {
	scQueryBlsKeys := &process.SCQuery{
		ScAddress: sc.AddressBytes(),
		FuncName:  "getBlsKeys",
		Arguments: [][]byte{},
	}

	vmOutputBlsKeys, err := dp.queryService.ExecuteQuery(scQueryBlsKeys)
	if err != nil {
		return err
	}
	delegatedNodes := dp.nodesListSplitter.GetDelegatedNodes(sc.AddressBytes())
	nodesAddresses := make([][]byte, 0, len(delegatedNodes))
	for _, node := range delegatedNodes {
		nodesAddresses = append(nodesAddresses, node.PubKeyBytes())
	}

	return dp.sameElements(vmOutputBlsKeys.ReturnData, nodesAddresses)
}

func (dp *delegationProcessor) sameElements(scReturned [][]byte, loaded [][]byte) error {
	if len(scReturned) != len(loaded) {
		return fmt.Errorf("%w staked nodes mismatch: %d found in SC, %d provided",
			genesis.ErrWhileVerifyingDelegation, len(scReturned), len(loaded))
	}

	sort.Slice(scReturned, func(i, j int) bool {
		return bytes.Compare(scReturned[i], scReturned[j]) < 0
	})
	sort.Slice(loaded, func(i, j int) bool {
		return bytes.Compare(loaded[i], loaded[j]) < 0
	})

	for i := 0; i < len(loaded); i++ {
		if !bytes.Equal(loaded[i], scReturned[i]) {
			return fmt.Errorf("%w, found in sc: %s, provided: %s",
				genesis.ErrMissingElement, hex.EncodeToString(scReturned[i]), hex.EncodeToString(loaded[i]))
		}
	}

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (dp *delegationProcessor) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
