package intermediate

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// codeMetadataHexForInitialSC used for initial SC deployment, set to upgrade-able
const codeMetadataHexForInitialSC = "0100"
const auctionScAddressPlaceholder = "%auction_sc_address%"
const versionFunction = "version"

// ArgDeployProcessor is the argument used to create a deployProcessor instance
type ArgDeployProcessor struct {
	Executor       genesis.TxExecutionProcessor
	PubkeyConv     core.PubkeyConverter
	BlockchainHook process.BlockChainHookHandler
	QueryService   external.SCQueryService
}

type deployProcessor struct {
	*baseDeploy
	pubkeyConv     core.PubkeyConverter
	scQueryService process.SCQueryService
}

// NewDeployProcessor returns a new instance of deploy processor able to deploy SC
func NewDeployProcessor(arg ArgDeployProcessor) (*deployProcessor, error) {
	if check.IfNil(arg.Executor) {
		return nil, genesis.ErrNilTxExecutionProcessor
	}
	if check.IfNil(arg.PubkeyConv) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.BlockchainHook) {
		return nil, process.ErrNilBlockChainHook
	}
	if check.IfNil(arg.QueryService) {
		return nil, genesis.ErrNilQueryService
	}

	base := &baseDeploy{
		TxExecutionProcessor: arg.Executor,
		pubkeyConv:           arg.PubkeyConv,
		blockchainHook:       arg.BlockchainHook,
		emptyAddress:         make([]byte, arg.PubkeyConv.Len()),
		getScCodeAsHex:       getSCCodeAsHex,
	}

	dp := &deployProcessor{
		pubkeyConv:     arg.PubkeyConv,
		scQueryService: arg.QueryService,
		baseDeploy:     base,
	}

	return dp, nil
}

// Deploy will try to deploy the provided smart contract
func (dp *deployProcessor) Deploy(sc genesis.InitialSmartContractHandler) ([][]byte, error) {
	code, err := dp.getScCodeAsHex(sc.GetFilename())
	if err != nil {
		return nil, err
	}

	scResultingAddressBytes, err := dp.deployForOneAddress(sc, sc.OwnerBytes(), code, applyCommonPlaceholders(sc.GetInitParameters()))
	if err != nil {
		return nil, err
	}

	return [][]byte{scResultingAddressBytes}, dp.checkVersion(sc, scResultingAddressBytes)
}

func applyCommonPlaceholders(txData string) string {
	//replace all placeholders containing auctionScAddressPlaceholder with the real hex address
	txData = strings.Replace(txData, auctionScAddressPlaceholder, hex.EncodeToString(vm.AuctionSCAddress), -1)

	return txData
}

func (dp *deployProcessor) checkVersion(sc genesis.InitialSmartContractHandler, scResultingAddressBytes []byte) error {
	if len(sc.GetVersion()) == 0 {
		//no version info, assuming deployed contract is up-to-date (let contracts that do not provide "version" function
		// to be deployed at genesis time)
		return nil
	}

	vc, err := versioning.NewVersionComparator(sc.GetVersion())
	if err != nil {
		return err
	}

	scQueryVersion := &process.SCQuery{
		ScAddress: scResultingAddressBytes,
		FuncName:  versionFunction,
		Arguments: [][]byte{},
	}

	vmOutputVersion, err := dp.scQueryService.ExecuteQuery(scQueryVersion)
	if err != nil {
		return err
	}
	if len(vmOutputVersion.ReturnData) != 1 {
		return genesis.ErrGetVersionFromSC
	}

	version := string(vmOutputVersion.ReturnData[0])

	log.Debug("SC version",
		"SC address", scResultingAddressBytes,
		"SC owner", sc.GetOwner(),
		"version", version,
	)

	return vc.Check(version)
}

// IsInterfaceNil returns if underlying object is true
func (dp *deployProcessor) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
