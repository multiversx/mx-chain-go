package intermediate

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/vm/factory"
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
	genesis.TxExecutionProcessor
	pubkeyConv     core.PubkeyConverter
	getScCodeAsHex func(filename string) (string, error)
	blockchainHook process.BlockChainHookHandler
	scQueryService process.SCQueryService
	emptyAddress   []byte
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

	dp := &deployProcessor{
		TxExecutionProcessor: arg.Executor,
		pubkeyConv:           arg.PubkeyConv,
		blockchainHook:       arg.BlockchainHook,
		scQueryService:       arg.QueryService,
	}
	dp.getScCodeAsHex = dp.getSCCodeAsHex
	dp.emptyAddress = make([]byte, dp.pubkeyConv.Len())

	return dp, nil
}

// Deploy will try to deploy the provided smart contract
func (dp *deployProcessor) Deploy(sc genesis.InitialSmartContractHandler) error {
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

	return dp.checkVersion(sc, scResultingAddressBytes)
}

func (dp *deployProcessor) applyCommonPlaceholders(txData string) string {
	//replace all placeholders containing auctionScAddressPlaceholder with the real hex address
	txData = strings.Replace(txData, auctionScAddressPlaceholder, hex.EncodeToString(vmFactory.AuctionSCAddress), -1)

	return txData
}

func (dp *deployProcessor) getSCCodeAsHex(filename string) (string, error) {
	code, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(code), nil
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
		"SC address", sc.Address(),
		"SC owner", sc.GetOwner(),
		"version", version,
	)

	return vc.Check(version)
}

// IsInterfaceNil returns if underlying object is true
func (dp *deployProcessor) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
