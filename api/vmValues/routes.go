package vmValues

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/gin-gonic/gin"
)

const (
	hexPath    = "/hex"
	stringPath = "/string"
	intPath    = "/int"
	queryPath  = "/query"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	ExecuteSCQuery(*process.SCQuery) (*vm.VMOutputApi, error)
	DecodeAddressPubkey(pk string) ([]byte, error)
	IsInterfaceNil() bool
}

// VMValueRequest represents the structure on which user input for generating a new transaction will validate against
type VMValueRequest struct {
	ScAddress  string   `form:"scAddress" json:"scAddress"`
	FuncName   string   `form:"funcName" json:"funcName"`
	CallerAddr string   `form:"caller" json:"caller"`
	CallValue  string   `form:"value" json:"value"`
	Args       []string `form:"args"  json:"args"`
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodPost, hexPath, getHex)
	router.RegisterHandler(http.MethodPost, stringPath, getString)
	router.RegisterHandler(http.MethodPost, intPath, getInt)
	router.RegisterHandler(http.MethodPost, queryPath, executeQuery)
}

// getHex returns the data as bytes, hex-encoded
func getHex(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsHex)
}

// getString returns the data as string
func getString(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsString)
}

// getInt returns the data as big int
func getInt(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsBigIntString)
}

func doGetVMValue(context *gin.Context, asType vmcommon.ReturnDataKind) {
	vmOutput, err := doExecuteQuery(context)

	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnData, err := vmOutput.GetFirstReturnData(asType)
	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnOkResponse(context, returnData)
}

// executeQuery returns the data as string
func executeQuery(context *gin.Context) {
	vmOutput, err := doExecuteQuery(context)
	if err != nil {
		returnBadRequest(context, "executeQuery", err)
		return
	}

	returnOkResponse(context, vmOutput)
}

func doExecuteQuery(context *gin.Context) (*vm.VMOutputApi, error) {
	efObj, ok := context.Get("facade")
	if !ok {
		return nil, errors.ErrNilAppContext
	}

	ef, ok := efObj.(FacadeHandler)
	if !ok {
		return nil, errors.ErrInvalidAppContext
	}

	request := VMValueRequest{}
	err := context.ShouldBindJSON(&request)
	if err != nil {
		return nil, errors.ErrInvalidJSONRequest
	}

	command, err := createSCQuery(ef, &request)
	if err != nil {
		return nil, err
	}

	return ef.ExecuteSCQuery(command)
}

func createSCQuery(fh FacadeHandler, request *VMValueRequest) (*process.SCQuery, error) {
	decodedAddress, err := fh.DecodeAddressPubkey(request.ScAddress)
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
		ScAddress: decodedAddress,
		FuncName:  request.FuncName,
		Arguments: arguments,
	}

	if len(request.CallerAddr) > 0 {
		callerAddress, errDecodeCaller := fh.DecodeAddressPubkey(request.CallerAddr)
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

func returnBadRequest(context *gin.Context, errScope string, err error) {
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

func returnOkResponse(context *gin.Context, data interface{}) {
	context.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"data": data},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
