package vmValues

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	SimulateRunSmartContractFunction(address []byte, funcName string, argsBuff ...[]byte) (interface{}, error)
	IsInterfaceNil() bool
}

// VMValueRequest represents the structure on which user input for generating a new transaction will validate against
type VMValueRequest struct {
	ScAddress string   `form:"scAddress" json:"scAddress"`
	FuncName  string   `form:"funcName" json:"funcName"`
	Args      []string `form:"args"  json:"args"`
}

type runFunctionCommand struct {
	ScAddress string
	FuncName  string
	Args      [][]byte // or big ints already?
}

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/hex", getHex)
	router.POST("/string", getString)
	router.POST("/int", getInt)
	router.POST("/simulate-run", SimulateRunFunction)
}

// getHex returns the data as bytes, hex-encoded
func getHex(context *gin.Context) {
	doGetVMValue(context, smartContract.AsHex)
}

// getString returns the data as string
func getString(context *gin.Context) {
	doGetVMValue(context, smartContract.AsString)
}

// getInt returns the data as big int
func getInt(context *gin.Context) {
	doGetVMValue(context, smartContract.AsBigInt)
}

func doGetVMValue(context *gin.Context, asType smartContract.ReturnDataAsType) {
	vmOutput, err := doSimulateRunFunction(context)

	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnData, err := smartContract.GetFirstReturnData(vmOutput, asType)
	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnOkResponse(context, returnData)
}

// SimulateRunFunction returns the data as string
func SimulateRunFunction(context *gin.Context) {
	vmOutput, err := doSimulateRunFunction(context)
	if err != nil {
		returnBadRequest(context, "SimulateRunSmartContractFunction", err)
		return
	}

	returnOkResponse(context, vmOutput)
}

func doSimulateRunFunction(context *gin.Context) (*vmcommon.VMOutput, error) {
	facade, ok := context.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		return nil, errors.ErrInvalidAppContext
	}

	var request = VMValueRequest{}
	err := context.ShouldBindJSON(&request)
	if err != nil {
		return nil, err
	}

	argsBuff := make([][]byte, 0)
	for _, arg := range request.Args {
		buff, err := hex.DecodeString(arg)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid hex string: %s", arg, err.Error())
		}

		argsBuff = append(argsBuff, buff)
	}

	adrBytes, err := hex.DecodeString(request.ScAddress)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not a valid hex string: %s", request.ScAddress, err.Error())
	}

	vmOutput, err := facade.SimulateRunSmartContractFunction(adrBytes, request.FuncName, argsBuff...)
	if err != nil {
		return nil, err
	}

	return vmOutput.(*vmcommon.VMOutput), nil
}

func returnBadRequest(context *gin.Context, errScope string, err error) {
	message := fmt.Sprintf("%s: %s", errScope, err)
	context.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func returnOkResponse(context *gin.Context, data interface{}) {
	context.JSON(http.StatusOK, gin.H{"data": data})
}
