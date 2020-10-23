package vm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// VMOutputApi is a wrapper over the vmcommon's VMOutput
type VMOutputApi struct {
	ReturnData      [][]byte                     `json:"returnData"`
	ReturnCode      string                       `json:"returnCode"`
	ReturnMessage   string                       `json:"returnMessage"`
	GasRemaining    uint64                       `json:"gasRemaining"`
	GasRefund       *big.Int                     `json:"gasRefund"`
	OutputAccounts  map[string]*OutputAccountApi `json:"outputAccounts"`
	DeletedAccounts [][]byte                     `json:"deletedAccounts"`
	TouchedAccounts [][]byte                     `json:"touchedAccounts"`
	Logs            []*LogEntryApi               `json:"logs"`
}

// StorageUpdateApi is a wrapper over vmcommon's StorageUpdate
type StorageUpdateApi struct {
	Offset []byte `json:"offset"`
	Data   []byte `json:"data"`
}

// OutputAccountApi is a wrapper over vmcommon's OutputAccount
type OutputAccountApi struct {
	Address         string                       `json:"address"`
	Nonce           uint64                       `json:"nonce"`
	Balance         *big.Int                     `json:"balance"`
	BalanceDelta    *big.Int                     `json:"balanceDelta"`
	StorageUpdates  map[string]*StorageUpdateApi `json:"storageUpdates"`
	Code            []byte                       `json:"code"`
	CodeMetadata    []byte                       `json:"codeMetaData"`
	OutputTransfers []OutputTransferApi          `json:"outputTransfers"`
	CallType        vmcommon.CallType            `json:"callType"`
}

// OutputTransferApi is a wrapper over vmcommon's OutputTransfer
type OutputTransferApi struct {
	Value    *big.Int          `json:"value"`
	GasLimit uint64            `json:"gasLimit"`
	Data     []byte            `json:"data"`
	CallType vmcommon.CallType `json:"callType"`
}

// LogEntryApi is a wrapper over vmcommon's LogEntry
type LogEntryApi struct {
	Identifier []byte   `json:"identifier"`
	Address    string   `json:"address"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
}

// GetFirstReturnData is a helper function that returns the first ReturnData of VMOutput, interpreted as specified.
func (vmOutput *VMOutputApi) GetFirstReturnData(asType vmcommon.ReturnDataKind) (interface{}, error) {
	if len(vmOutput.ReturnData) == 0 {
		return nil, fmt.Errorf("no return data")
	}

	returnData := vmOutput.ReturnData[0]

	switch asType {
	case vmcommon.AsBigInt:
		return big.NewInt(0).SetBytes(returnData), nil
	case vmcommon.AsBigIntString:
		return big.NewInt(0).SetBytes(returnData).String(), nil
	case vmcommon.AsString:
		return string(returnData), nil
	case vmcommon.AsHex:
		return hex.EncodeToString(returnData), nil
	}

	return nil, fmt.Errorf("can't interpret return data")
}
