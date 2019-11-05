package vmValues

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	GetVmOutput(address string, funcName string, argsBuff ...[]byte) (*vmcommon.VMOutput, error)
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
	router.POST("/vm-output", GetVMOutput)
}

func doGetVMOutput(context *gin.Context) (*vmcommon.VMOutput, error) {
	ef, _ := context.MustGet("elrondFacade").(FacadeHandler)

	var gval = VMValueRequest{}
	err := context.ShouldBindJSON(&gval)
	if err != nil {
		return nil, err
	}

	argsBuff := make([][]byte, 0)
	for _, arg := range gval.Args {
		buff, err := hex.DecodeString(arg)
		if err != nil {
			return nil,
				errors.New(fmt.Sprintf("'%s' is not a valid hex string: %s", arg, err.Error()))
		}

		argsBuff = append(argsBuff, buff)
	}

	adrBytes, err := hex.DecodeString(gval.ScAddress)
	if err != nil {
		return nil,
			errors.New(fmt.Sprintf("'%s' is not a valid hex string: %s", gval.ScAddress, err.Error()))
	}

	vmOutput, err := ef.GetVmOutput(string(adrBytes), gval.FuncName, argsBuff...)
	if err != nil {
		return nil, err
	}

	return vmOutput, nil
}

func doGetVMReturnData(context *gin.Context) ([]byte, error) {
	vmOutput, err := doGetVMOutput(context)
	if err != nil {
		return nil, err
	}

	if len(vmOutput.ReturnData) == 0 {
		return nil, fmt.Errorf("no return data")
	}

	returnData := vmOutput.ReturnData[0].Bytes()
	return returnData, nil
}

// GetVMValueAsHexBytes returns the data as byte slice
func GetVMValueAsHexBytes(context *gin.Context) {
	data, err := doGetVMReturnData(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsHexBytes: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": hex.EncodeToString(data)})
}

// GetVMValueAsString returns the data as string
func GetVMValueAsString(context *gin.Context) {
	data, err := doGetVMReturnData(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsString: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": string(data)})
}

// GetVMValueAsBigInt returns the data as big int
func GetVMValueAsBigInt(context *gin.Context) {
	data, err := doGetVMReturnData(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMValueAsBigInt: %s", err)})
		return
	}

	value := big.NewInt(0).SetBytes(data)
	context.JSON(http.StatusOK, gin.H{"data": value.String()})
}

// GetVMOutput returns the data as string
func GetVMOutput(context *gin.Context) {
	vmOutput, err := doGetVMOutput(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("GetVMOutput: %s", err)})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": vmOutput})
}
