package intermediate

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/vm/factory"
)

// codeMetadataHexForInitialSC used for initial SC deployment, set to upgrade-able
const codeMetadataHexForInitialSC = "0100"
const auctionScAddressPlaceholder = "%auction_sc_address%"

type deployProcessor struct {
	genesis.TxExecutionProcessor
	pubkeyConv          state.PubkeyConverter
	replacePlaceholders func(txData string, scResultingAddressBytes []byte) (string, error)
	getScCodeAsHex      func(filename string) (string, error)
	blockchainHook      process.BlockChainHookHandler
	emptyAddress        []byte
}

// NewDeployProcessor returns a new instance of deploy processor able to deploy SC
func NewDeployProcessor(
	executor genesis.TxExecutionProcessor,
	pubkeyConv state.PubkeyConverter,
	blockchainHook process.BlockChainHookHandler,
) (*deployProcessor, error) {
	if check.IfNil(executor) {
		return nil, genesis.ErrNilTxExecutionProcessor
	}
	if check.IfNil(pubkeyConv) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(blockchainHook) {
		return nil, process.ErrNilBlockChainHook
	}

	dp := &deployProcessor{
		TxExecutionProcessor: executor,
		pubkeyConv:           pubkeyConv,
		blockchainHook:       blockchainHook,
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

	vmType := sc.GetVmType()
	deployTxData := strings.Join([]string{code, vmType, codeMetadataHexForInitialSC}, "@")
	deployTxData = dp.applyCommonPlaceholders(deployTxData)
	if dp.replacePlaceholders != nil {
		deployTxData, err = dp.replacePlaceholders(deployTxData, scResultingAddressBytes)
		if err != nil {
			return err
		}
	}

	return dp.ExecuteTransaction(
		nonce,
		sc.OwnerBytes(),
		dp.emptyAddress,
		big.NewInt(0),
		[]byte(deployTxData),
	)
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
