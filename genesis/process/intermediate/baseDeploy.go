package intermediate

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
)

type baseDeploy struct {
	genesis.TxExecutionProcessor
	pubkeyConv     core.PubkeyConverter
	blockchainHook process.BlockChainHookHandler
	emptyAddress   []byte
	getScCodeAsHex func(filename string) (string, error)
}

func (dp *baseDeploy) deployForOneAddress(
	sc genesis.InitialSmartContractHandler,
	ownerAddressBytes []byte,
	code string,
	initParams string,
) ([]byte, error) {
	nonce, err := dp.GetNonce(ownerAddressBytes)
	if err != nil {
		return nil, err
	}

	scResultingAddressBytes, err := dp.blockchainHook.NewAddress(
		ownerAddressBytes,
		nonce,
		sc.VmTypeBytes(),
	)
	if err != nil {
		return nil, err
	}

	sc.AddAddressBytes(scResultingAddressBytes)

	scResultingAddress, err := dp.pubkeyConv.Encode(scResultingAddressBytes)
	if err != nil {
		return nil, err
	}
	sc.AddAddress(scResultingAddress)

	vmType := sc.GetVmType()
	arguments := []string{code, vmType, codeMetadataHexForInitialSC}
	if len(initParams) > 0 {
		arguments = append(arguments, initParams)
	}
	deployTxData := strings.Join(arguments, "@")

	log.Trace("deploying genesis SC",
		"SC owner", sc.GetOwner(),
		"SC ownerAddress", scResultingAddress,
		"type", sc.GetType(),
		"VM type", sc.GetVmType(),
		"init params", initParams,
	)

	_, accountExists := dp.GetAccount(scResultingAddressBytes)
	if accountExists {
		return nil, fmt.Errorf("%w for SC ownerAddress %s, owner %s with nonce %d",
			genesis.ErrAccountAlreadyExists,
			scResultingAddressBytes,
			ownerAddressBytes,
			nonce,
		)
	}

	err = dp.ExecuteTransaction(
		nonce,
		ownerAddressBytes,
		dp.emptyAddress,
		big.NewInt(0),
		[]byte(deployTxData),
	)
	if err != nil {
		return nil, err
	}

	ownerAddress, err := dp.pubkeyConv.Encode(ownerAddressBytes)
	if err != nil {
		return nil, err
	}

	_, accountExists = dp.GetAccount(scResultingAddressBytes)
	if !accountExists {
		return nil, fmt.Errorf("%w for SC ownerAddress %s, owner %s with nonce %d",
			genesis.ErrAccountNotCreated,
			scResultingAddress,
			ownerAddress,
			nonce,
		)
	}

	return scResultingAddressBytes, nil
}

func getSCCodeAsHex(filename string) (string, error) {
	code, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(code), nil
}
