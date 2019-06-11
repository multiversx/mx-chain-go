package getValues

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	apiErrors "github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	GetDataValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
}

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/as-hex", GetDataValueAsHexBytes)
	router.POST("/as-string", GetDataValueAsString)
	router.POST("/as-int", GetDataValueAsBigInt)
}

func getDataValueFromAccount(c *gin.Context) ([]byte, int, error) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		return nil, http.StatusInternalServerError, apiErrors.ErrInvalidAppContext
	}

	addr := c.Param("address")
	if addr == "" {
		return nil, http.StatusBadRequest, apiErrors.ErrEmptyAddress
	}

	funcName := c.Param("func")
	if funcName == "" {
		return nil, http.StatusBadRequest, apiErrors.ErrEmptyFuncName
	}

	argsBuff := make([][]byte, 0)
	argsString, exists := c.GetPostFormArray("args")
	if exists {
		for _, arg := range argsString {
			buff, err := hex.DecodeString(arg)
			if err != nil {
				return nil,
					http.StatusBadRequest,
					errors.New(fmt.Sprintf("'%s' is not a valid hex string: %s", arg, err.Error()))
			}

			argsBuff = append(argsBuff, buff)
		}
	}

	returnedData, err := ef.GetDataValue(addr, funcName, argsBuff...)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return returnedData, http.StatusOK, nil
}

// GetDataValueAsHexBytes returns the data as byte slice
func GetDataValueAsHexBytes(c *gin.Context) {
	data, status, err := getDataValueFromAccount(c)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get value as hex bytes: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hex.EncodeToString(data)})
}

// GetDataValueAsString returns the data as string
func GetDataValueAsString(c *gin.Context) {
	data, status, err := getDataValueFromAccount(c)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get value as string: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": string(data)})
}

// GetDataValueAsBigInt returns the data as big int
func GetDataValueAsBigInt(c *gin.Context) {
	data, status, err := getDataValueFromAccount(c)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get value as big int: %s", err)})
		return
	}

	value := big.NewInt(0).SetBytes(data)
	c.JSON(http.StatusOK, gin.H{"data": value.String()})
}
