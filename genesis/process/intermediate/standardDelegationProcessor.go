package intermediate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgStandardDelegationProcessor is the argument used to construct a standard delegation processor
type ArgStandardDelegationProcessor struct {
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

type standardDelegationProcessor struct {
	genesis.TxExecutionProcessor
	shardCoordinator     sharding.Coordinator
	accuntsParser        genesis.AccountsParser
	smartContractsParser genesis.InitialSmartContractParser
	nodesListSplitter    genesis.NodesListSplitter
	queryService         external.SCQueryService
	nodePrice            *big.Int
}

// NewStandardDelegationProcessor returns a new standard delegation processor instance
func NewStandardDelegationProcessor(arg ArgStandardDelegationProcessor) (*standardDelegationProcessor, error) {
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

	return &standardDelegationProcessor{
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
func (sdp *standardDelegationProcessor) ExecuteDelegation() (genesis.DelegationResult, error) {
	smartContracts, err := sdp.getDelegationScOnCurrentShard()
	if err != nil {
		return genesis.DelegationResult{}, err
	}
	if len(smartContracts) == 0 {
		return genesis.DelegationResult{}, nil
	}

	err = sdp.setDelegationStartParameters(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	_, err = sdp.executeManageBlsKeys(smartContracts, sdp.getBlsKey, setBlsKeysFunction)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	dr := genesis.DelegationResult{}
	dr.NumTotalStaked, err = sdp.executeStake(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	dr.NumTotalDelegated, err = sdp.executeManageBlsKeys(smartContracts, sdp.getBlsKeySig, activateBlsKeysFunction)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	err = sdp.executeVerify(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	return dr, err
}

func (sdp *standardDelegationProcessor) getDelegationScOnCurrentShard() ([]genesis.InitialSmartContractHandler, error) {
	allSmartContracts, err := sdp.smartContractsParser.InitialSmartContractsSplitOnOwnersShards(sdp.shardCoordinator)
	if err != nil {
		return nil, err
	}

	smartContracts := make([]genesis.InitialSmartContractHandler, 0)
	smartContractsForCurrentShard := allSmartContracts[sdp.shardCoordinator.SelfId()]
	for _, sc := range smartContractsForCurrentShard {
		if sc.GetType() == genesis.DelegationType {
			smartContracts = append(smartContracts, sc)
		}
	}

	log.Trace("getDelegationScOnCurrentShard",
		"num delegation SC", len(smartContracts),
		"shard ID", sdp.shardCoordinator.SelfId(),
	)
	return smartContracts, nil
}

func getDeployedSCAddress(sc genesis.InitialSmartContractHandler) string {
	if len(sc.Addresses()) != 1 {
		return ""
	}
	return sc.Addresses()[0]
}

func getDeployedSCAddressBytes(sc genesis.InitialSmartContractHandler) []byte {
	if len(sc.AddressesBytes()) != 1 {
		return nil
	}
	return sc.AddressesBytes()[0]
}

func (sdp *standardDelegationProcessor) setDelegationStartParameters(smartContracts []genesis.InitialSmartContractHandler) error {
	for _, sc := range smartContracts {

		delegatedNodes := sdp.nodesListSplitter.GetDelegatedNodes(getDeployedSCAddressBytes(sc))
		numNodes := len(delegatedNodes)

		log.Trace("setDelegationStartParameters",
			"SC owner", sc.GetOwner(),
			"SC address", getDeployedSCAddress(sc),
			"num delegated nodes", numNodes,
			"node price", sdp.nodePrice.String(),
			"shard ID", sdp.shardCoordinator.SelfId(),
		)

		err := sdp.executeSetNumNodes(numNodes, sc)
		if err != nil {
			return err
		}

		err = sdp.executeSetNodePrice(sc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sdp *standardDelegationProcessor) executeSetNumNodes(numNodes int, sc genesis.InitialSmartContractHandler) error {
	setNumNodesTxData := fmt.Sprintf("%s@%s", setNumNodesFunction, core.ConvertToEvenHex(numNodes))

	nonce, err := sdp.GetNonce(sc.OwnerBytes())
	if err != nil {
		return err
	}

	return sdp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		getDeployedSCAddressBytes(sc),
		zero,
		[]byte(setNumNodesTxData),
	)
}

func (sdp *standardDelegationProcessor) executeSetNodePrice(sc genesis.InitialSmartContractHandler) error {
	setStakePerNodeTxData := fmt.Sprintf("%s@%s", setStakePerNodeFunction, core.ConvertToEvenHexBigInt(sdp.nodePrice))

	nonce, err := sdp.GetNonce(sc.OwnerBytes())
	if err != nil {
		return err
	}

	return sdp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		getDeployedSCAddressBytes(sc),
		zero,
		[]byte(setStakePerNodeTxData),
	)
}

func (sdp *standardDelegationProcessor) executeStake(smartContracts []genesis.InitialSmartContractHandler) (int, error) {
	stakedOnDelegation := 0

	for _, sc := range smartContracts {
		accounts := sdp.accuntsParser.GetInitialAccountsForDelegated(getDeployedSCAddressBytes(sc))
		if len(accounts) == 0 {
			log.Debug("genesis delegation SC was not delegated by any account",
				"SC owner", sc.GetOwner(),
				"SC address", getDeployedSCAddress(sc),
			)
			continue
		}

		totalDelegated := big.NewInt(0)
		for _, ac := range accounts {
			err := sdp.stake(ac, sc)
			if err != nil {
				return 0, fmt.Errorf("%w while calling stake function from account %s", err, ac.GetAddress())
			}

			totalDelegated.Add(totalDelegated, ac.GetDelegationHandler().GetValue())
		}

		log.Trace("executeStake",
			"SC owner", sc.GetOwner(),
			"SC address", getDeployedSCAddress(sc),
			"num accounts", len(accounts),
			"total delegated", totalDelegated,
		)
		stakedOnDelegation += len(accounts)
	}

	return stakedOnDelegation, nil
}

func (sdp *standardDelegationProcessor) stake(ac genesis.InitialAccountHandler, sc genesis.InitialSmartContractHandler) error {
	isIntraShardCall := sdp.shardCoordinator.SameShard(ac.AddressBytes(), getDeployedSCAddressBytes(sc))

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
		nonce, err = sdp.GetNonce(ac.AddressBytes())
		if err != nil {
			return err
		}
	}

	stakeData := fmt.Sprintf("%s@%s", stakeFunction, dh.GetValue().Text(16))
	err = sdp.ExecuteTransaction(
		nonce,
		ac.AddressBytes(),
		getDeployedSCAddressBytes(sc),
		zero,
		[]byte(stakeData),
	)
	if err != nil {
		return err
	}

	return nil
}

func (sdp *standardDelegationProcessor) executeManageBlsKeys(
	smartContracts []genesis.InitialSmartContractHandler,
	handler func(node sharding.GenesisNodeInfoHandler) string,
	function string,
) (int, error) {

	log.Trace("executeManageSetBlsKeys",
		"num delegation SC", len(smartContracts),
		"shard ID", sdp.shardCoordinator.SelfId(),
		"function", function,
	)

	totalDelegated := 0
	for _, sc := range smartContracts {
		delegatedNodes := sdp.nodesListSplitter.GetDelegatedNodes(getDeployedSCAddressBytes(sc))

		lenDelegated := len(delegatedNodes)
		if lenDelegated == 0 {
			log.Debug("genesis delegation SC does not have staked nodes",
				"SC owner", sc.GetOwner(),
				"SC address", getDeployedSCAddress(sc),
				"function", function,
			)
			continue
		}
		totalDelegated += lenDelegated

		log.Trace("executeSetBlsKeys",
			"SC owner", sc.GetOwner(),
			"SC address", getDeployedSCAddress(sc),
			"num nodes", lenDelegated,
			"shard ID", sdp.shardCoordinator.SelfId(),
			"function", function,
		)

		arguments := make([]string, 0, len(delegatedNodes)+1)
		arguments = append(arguments, function)
		for _, node := range delegatedNodes {
			arg := handler(node)
			arguments = append(arguments, arg)
		}

		nonce, err := sdp.GetNonce(sc.OwnerBytes())
		if err != nil {
			return 0, err
		}

		err = sdp.ExecuteTransaction(
			nonce,
			sc.OwnerBytes(),
			getDeployedSCAddressBytes(sc),
			big.NewInt(0),
			[]byte(strings.Join(arguments, "@")),
		)
		if err != nil {
			return 0, err
		}
	}

	return totalDelegated, nil
}

func (sdp *standardDelegationProcessor) getBlsKey(node sharding.GenesisNodeInfoHandler) string {
	return hex.EncodeToString(node.PubKeyBytes())
}

func (sdp *standardDelegationProcessor) getBlsKeySig(_ sharding.GenesisNodeInfoHandler) string {
	mockSignature := []byte("genesis signature")

	return hex.EncodeToString(mockSignature)
}

func (sdp *standardDelegationProcessor) executeVerify(smartContracts []genesis.InitialSmartContractHandler) error {
	for _, sc := range smartContracts {
		err := sdp.verify(sc)
		if err != nil {
			return fmt.Errorf("%w for contract %s, owner %s", err, getDeployedSCAddress(sc), sc.GetOwner())
		}
	}

	return nil
}

func (sdp *standardDelegationProcessor) verify(sc genesis.InitialSmartContractHandler) error {
	err := sdp.verifyStakedValue(sc)
	if err != nil {
		return fmt.Errorf("%w for verifyStakedValue", err)
	}

	err = sdp.verifyRegisteredNodes(sc)
	if err != nil {
		return fmt.Errorf("%w for verifyRegisteredNodes", err)
	}

	return nil
}

func (sdp *standardDelegationProcessor) verifyStakedValue(sc genesis.InitialSmartContractHandler) error {
	scQueryStakeValue := &process.SCQuery{
		ScAddress: getDeployedSCAddressBytes(sc),
		FuncName:  "getFilledStake",
		Arguments: [][]byte{},
	}
	vmOutputStakeValue, err := sdp.queryService.ExecuteQuery(scQueryStakeValue)
	if err != nil {
		return err
	}
	if len(vmOutputStakeValue.ReturnData) != 1 {
		return fmt.Errorf("%w return data should have contained one element", genesis.ErrWhileVerifyingDelegation)
	}
	scStakedValue := big.NewInt(0).SetBytes(vmOutputStakeValue.ReturnData[0])
	providedStakedValue := big.NewInt(0)
	providedDelegators := sdp.accuntsParser.GetInitialAccountsForDelegated(getDeployedSCAddressBytes(sc))

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

func (sdp *standardDelegationProcessor) verifyRegisteredNodes(sc genesis.InitialSmartContractHandler) error {
	scQueryBlsKeys := &process.SCQuery{
		ScAddress: getDeployedSCAddressBytes(sc),
		FuncName:  "getBlsKeys",
		Arguments: [][]byte{},
	}

	vmOutputBlsKeys, err := sdp.queryService.ExecuteQuery(scQueryBlsKeys)
	if err != nil {
		return err
	}
	delegatedNodes := sdp.nodesListSplitter.GetDelegatedNodes(getDeployedSCAddressBytes(sc))
	nodesAddresses := make([][]byte, 0, len(delegatedNodes))
	for _, node := range delegatedNodes {
		nodesAddresses = append(nodesAddresses, node.PubKeyBytes())
	}

	return sdp.sameElements(vmOutputBlsKeys.ReturnData, nodesAddresses)
}

func (sdp *standardDelegationProcessor) sameElements(scReturned [][]byte, loaded [][]byte) error {
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
func (sdp *standardDelegationProcessor) IsInterfaceNil() bool {
	return sdp == nil || sdp.TxExecutionProcessor == nil
}
