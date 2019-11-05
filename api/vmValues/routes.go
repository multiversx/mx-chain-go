package vmValues

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	SimulateRunSmartContractFunction(address string, funcName string, argsBuff ...[]byte) (interface{}, error)
	IsInterfaceNil() bool
}

// VMValueRequest represents the structure on which user input for generating a new transaction will validate against
type VMValueRequest struct {
	ScAddress string   `form:"scAddress" json:"scAddress"`
	FuncName  string   `form:"funcName" json:"funcName"`
	Args      []string `form:"args"  json:"args"`
}

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/hex", GetVMValueAsHexBytes)
	router.POST("/string", GetVMValueAsString)
	router.POST("/int", GetVMValueAsBigInt)
	router.POST("/simulate-run", SimulateRunFunction)
}

// GetVMValueAsHexBytes returns the data as byte slice
func GetVMValueAsHexBytes(context *gin.Context) {
	data, err := doSimulateRunFunctionAndGetFirstOutput(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsHexBytes: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": hex.EncodeToString(data)})
}

// GetVMValueAsString returns the data as string
func GetVMValueAsString(context *gin.Context) {
	data, err := doSimulateRunFunctionAndGetFirstOutput(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsString: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": string(data)})
}

// GetVMValueAsBigInt returns the data as big int
func GetVMValueAsBigInt(context *gin.Context) {
	data, err := doSimulateRunFunctionAndGetFirstOutput(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsBigInt: %s", err)})
		return
	}

	value := big.NewInt(0).SetBytes(data)
	context.JSON(http.StatusOK, gin.H{"data": value.String()})
}

// SimulateRunFunction returns the data as string
func SimulateRunFunction(context *gin.Context) {
	vmOutput, err := doSimulateRunFunction(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("SimulateRunSmartContractFunction: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": vmOutput})
}

func doSimulateRunFunctionAndGetFirstOutput(context *gin.Context) ([]byte, error) {
	vmOutput, err := doSimulateRunFunction(context)
	if err != nil {
		return nil, err
	}

	if len(vmOutput.ReturnData) == 0 {
		return nil, fmt.Errorf("no return data")
	}

	returnData := vmOutput.ReturnData[0].Bytes()
	return returnData, nil
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

	vmOutput, err := facade.SimulateRunSmartContractFunction(string(adrBytes), request.FuncName, argsBuff...)
	if err != nil {
		return nil, err
	}

	return vmOutput.(*vmcommon.VMOutput), nil
}
