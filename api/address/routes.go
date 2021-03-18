package address

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-gonic/gin"
)

const (
	getAccountPath  = "/:address"
	getBalancePath  = "/:address/balance"
	getUsernamePath = "/:address/username"
	getKeysPath     = "/:address/keys"
	getKeyPath      = "/:address/key/:key"
	getESDTTokens   = "/:address/esdt"
	getESDTBalance  = "/:address/esdt/:tokenIdentifier"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	GetBalance(address string) (*big.Int, error)
	GetUsername(address string) (string, error)
	GetValueForKey(address string, key string) (string, error)
	GetAccount(address string) (state.UserAccountHandler, error)
	GetCode(account state.UserAccountHandler) []byte
	GetESDTBalance(address string, key string) (string, string, error)
	GetAllESDTTokens(address string) ([]string, error)
	GetKeyValuePairs(address string) (map[string]string, error)
	IsInterfaceNil() bool
}

type accountResponse struct {
	Address  string `json:"address"`
	Nonce    uint64 `json:"nonce"`
	Balance  string `json:"balance"`
	Username string `json:"username"`
	Code     string `json:"code"`
	CodeHash []byte `json:"codeHash"`
	RootHash []byte `json:"rootHash"`
}

type esdtTokenData struct {
	TokenIdentifier string `json:"tokenIdentifier"`
	Balance         string `json:"balance"`
	Properties      string `json:"properties"`
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, getAccountPath, GetAccount)
	router.RegisterHandler(http.MethodGet, getBalancePath, GetBalance)
	router.RegisterHandler(http.MethodGet, getUsernamePath, GetUsername)
	router.RegisterHandler(http.MethodGet, getKeyPath, GetValueForKey)
	router.RegisterHandler(http.MethodGet, getKeysPath, GetKeyValuePairs)
	router.RegisterHandler(http.MethodGet, getESDTBalance, GetESDTBalance)
	router.RegisterHandler(http.MethodGet, getESDTTokens, GetESDTTokens)
}

func getFacade(c *gin.Context) (FacadeHandler, bool) {
	facadeObj, ok := c.Get("facade")
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNilAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	facade, ok := facadeObj.(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	return facade, true
}

// GetAccount returns an accountResponse containing information
//  about the account correlated with provided address
func GetAccount(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	acc, err := facade.GetAccount(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrCouldNotGetAccount.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	code := facade.GetCode(acc)
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"account": accountResponseFromBaseAccount(addr, code, acc)},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetBalance returns the balance for the address parameter
func GetBalance(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	balance, err := facade.GetBalance(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"balance": balance.String()},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetUsername returns the username for the address parameter
func GetUsername(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetUsername.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	userName, err := facade.GetUsername(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetUsername.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"username": userName},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetValueForKey returns the value for the given address and key
func GetValueForKey(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyKey.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	value, err := facade.GetValueForKey(addr, key)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"value": value},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetKeyValuePairs returns all the key-value pairs for the given address
func GetKeyValuePairs(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetKeyValuePairs.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	value, err := facade.GetKeyValuePairs(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetKeyValuePairs.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"pairs": value},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetESDTBalance returns the balance for the given address and esdt token
func GetESDTBalance(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokenIdentifier := c.Param("tokenIdentifier")
	if tokenIdentifier == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), errors.ErrEmptyKey.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	balance, freeze, err := facade.GetESDTBalance(addr, tokenIdentifier)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTBalance.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	tokenData := esdtTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         balance,
		Properties:      freeze,
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokenData": tokenData},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetESDTTokens returns the tokens list from this account
func GetESDTTokens(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTTokens.Error(), errors.ErrEmptyAddress.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tokens, err := facade.GetAllESDTTokens(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetESDTTokens.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"tokens": tokens},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func accountResponseFromBaseAccount(address string, code []byte, account state.UserAccountHandler) accountResponse {
	return accountResponse{
		Address:  address,
		Nonce:    account.GetNonce(),
		Balance:  account.GetBalance().String(),
		Username: string(account.GetUserName()),
		Code:     hex.EncodeToString(code),
		CodeHash: account.GetCodeHash(),
		RootHash: account.GetRootHash(),
	}
}
