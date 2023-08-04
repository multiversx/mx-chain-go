package groups

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	hexPath    = "/hex"
	stringPath = "/string"
	intPath    = "/int"
	queryPath  = "/query"
)

// vmValuesFacadeHandler defines the methods to be implemented by a facade for vm-values requests
type vmValuesFacadeHandler interface {
	ExecuteSCQuery(*process.SCQuery) (*vm.VMOutputApi, error)
	DecodeAddressPubkey(pk string) ([]byte, error)
	IsInterfaceNil() bool
}

type vmValuesGroup struct {
	*baseGroup
	facade    vmValuesFacadeHandler
	mutFacade sync.RWMutex
}

// NewVmValuesGroup returns a new instance of vmValuesGroup
func NewVmValuesGroup(facade vmValuesFacadeHandler) (*vmValuesGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for vm values group", errors.ErrNilFacadeHandler)
	}

	vvg := &vmValuesGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    hexPath,
			Method:  http.MethodPost,
			Handler: vvg.getHex,
		},
		{
			Path:    stringPath,
			Method:  http.MethodPost,
			Handler: vvg.getString,
		},
		{
			Path:    intPath,
			Method:  http.MethodPost,
			Handler: vvg.getInt,
		},
		{
			Path:    queryPath,
			Method:  http.MethodPost,
			Handler: vvg.executeQuery,
		},
	}
	vvg.endpoints = endpoints

	return vvg, nil
}

// VMValueRequest represents the structure on which user input for generating a new transaction will validate against
type VMValueRequest struct {
	ScAddress      string   `json:"scAddress"`
	FuncName       string   `json:"funcName"`
	CallerAddr     string   `json:"caller"`
	CallValue      string   `json:"value"`
	Args           []string `json:"args"`
	SameScState    bool     `json:"sameScState"`
	ShouldBeSynced bool     `json:"shouldBeSynced"`
}

// getHex returns the data as bytes, hex-encoded
func (vvg *vmValuesGroup) getHex(context *gin.Context) {
	vvg.doGetVMValue(context, vm.AsHex)
}

// getString returns the data as string
func (vvg *vmValuesGroup) getString(context *gin.Context) {
	vvg.doGetVMValue(context, vm.AsString)
}

// getInt returns the data as big int
func (vvg *vmValuesGroup) getInt(context *gin.Context) {
	vvg.doGetVMValue(context, vm.AsBigIntString)
}

func (vvg *vmValuesGroup) doGetVMValue(context *gin.Context, asType vm.ReturnDataKind) {
	vmOutput, execErrMsg, err := vvg.doExecuteQuery(context)

	if err != nil {
		vvg.returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnData, err := vmOutput.GetFirstReturnData(asType)
	if err != nil {
		execErrMsg += " " + err.Error()
	}

	vvg.returnOkResponse(context, returnData, execErrMsg)
}

// executeQuery returns the data as string
func (vvg *vmValuesGroup) executeQuery(context *gin.Context) {
	vmOutput, execErrMsg, err := vvg.doExecuteQuery(context)
	if err != nil {
		vvg.returnBadRequest(context, "executeQuery", err)
		return
	}

	vvg.returnOkResponse(context, vmOutput, execErrMsg)
}

func (vvg *vmValuesGroup) doExecuteQuery(context *gin.Context) (*vm.VMOutputApi, string, error) {
	request := VMValueRequest{}
	err := context.ShouldBindJSON(&request)
	if err != nil {
		return nil, "", errors.ErrInvalidJSONRequest
	}

	command, err := vvg.createSCQuery(&request)
	if err != nil {
		return nil, "", err
	}

	command.BlockNonce, command.BlockHash, err = extractBlockCoordinates(context)
	if err != nil {
		return nil, "", err
	}

	vmOutputApi, err := vvg.getFacade().ExecuteSCQuery(command)
	if err != nil {
		return nil, "", err
	}

	vmExecErrMsg := ""
	if len(vmOutputApi.ReturnCode) > 0 && vmOutputApi.ReturnCode != vmcommon.Ok.String() {
		vmExecErrMsg = vmOutputApi.ReturnCode + ":" + vmOutputApi.ReturnMessage
	}

	return vmOutputApi, vmExecErrMsg, nil
}

func extractBlockCoordinates(context *gin.Context) (core.OptionalUint64, []byte, error) {
	blockNonce, err := parseUint64UrlParam(context, urlParamBlockNonce)
	if err != nil {
		return core.OptionalUint64{}, nil, fmt.Errorf("%w for block nonce", err)
	}

	blockHash, err := parseHexBytesUrlParam(context, urlParamBlockHash)
	if err != nil {
		return core.OptionalUint64{}, nil, fmt.Errorf("%w for block hash", err)
	}

	return blockNonce, blockHash, nil
}

func (vvg *vmValuesGroup) createSCQuery(request *VMValueRequest) (*process.SCQuery, error) {
	decodedAddress, err := vvg.getFacade().DecodeAddressPubkey(request.ScAddress)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not a valid address: %s", request.ScAddress, err.Error())
	}

	arguments := make([][]byte, len(request.Args))
	var argBytes []byte
	for i, arg := range request.Args {
		argBytes, err = hex.DecodeString(arg)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid hex string: %s", arg, err.Error())
		}

		arguments[i] = append(arguments[i], argBytes...)
	}

	scQuery := &process.SCQuery{
		ScAddress:      decodedAddress,
		FuncName:       request.FuncName,
		Arguments:      arguments,
		SameScState:    request.SameScState,
		ShouldBeSynced: request.ShouldBeSynced,
	}

	if len(request.CallerAddr) > 0 {
		callerAddress, errDecodeCaller := vvg.getFacade().DecodeAddressPubkey(request.CallerAddr)
		if errDecodeCaller != nil {
			return nil, errDecodeCaller
		}

		scQuery.CallerAddr = callerAddress
	}

	if len(request.CallValue) > 0 {
		callValue, ok := big.NewInt(0).SetString(request.CallValue, 10)
		if !ok {
			return nil, fmt.Errorf("non numeric call value provided: %s", request.CallValue)
		}
		scQuery.CallValue = callValue
	}

	return scQuery, nil
}

func (vvg *vmValuesGroup) returnBadRequest(context *gin.Context, errScope string, err error) {
	message := fmt.Sprintf("%s: %s", errScope, err)
	context.JSON(
		http.StatusBadRequest,
		shared.GenericAPIResponse{
			Data:  nil,
			Error: message,
			Code:  shared.ReturnCodeRequestError,
		},
	)
}

func (vvg *vmValuesGroup) returnOkResponse(context *gin.Context, data interface{}, errorMsg string) {
	context.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"data": data},
			Error: errorMsg,
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (vvg *vmValuesGroup) getFacade() vmValuesFacadeHandler {
	vvg.mutFacade.RLock()
	defer vvg.mutFacade.RUnlock()

	return vvg.facade
}

// UpdateFacade will update the facade
func (vvg *vmValuesGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(vmValuesFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	vvg.mutFacade.Lock()
	vvg.facade = castFacade
	vvg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vvg *vmValuesGroup) IsInterfaceNil() bool {
	return vvg == nil
}
