package address

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	GetBalance(address string) (*big.Int, error)
	GetValueForKey(address string, key string) (string, error)
	GetAccount(address string) (state.UserAccountHandler, error)
	IsInterfaceNil() bool
}

type accountResponse struct {
	Address  string `json:"address"`
	Nonce    uint64 `json:"nonce"`
	Balance  string `json:"balance"`
	Code     string `json:"code"`
	CodeHash []byte `json:"codeHash"`
	RootHash []byte `json:"rootHash"`
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, "/:address", GetAccount)
	router.RegisterHandler(http.MethodGet, "/:address/balance", GetBalance)
	router.RegisterHandler(http.MethodGet, "/:address/key/:key", GetValueForKey)
}

// GetAccount returns an accountResponse containing information
//  about the account correlated with provided address
func GetAccount(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	addr := c.Param("address")
	acc, err := ef.GetAccount(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrCouldNotGetAccount.Error(), err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": accountResponseFromBaseAccount(addr, acc)})
}

// GetBalance returns the balance for the address parameter
func GetBalance(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}
	addr := c.Param("address")

	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), errors.ErrEmptyAddress.Error())})
		return
	}

	balance, err := ef.GetBalance(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance.String()})
}

// GetValueForKey returns the value for the given address and key
func GetValueForKey(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	addr := c.Param("address")
	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyAddress.Error())})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), errors.ErrEmptyKey.Error())})
		return
	}

	value, err := ef.GetValueForKey(addr, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrGetValueForKey.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": value})
}

func accountResponseFromBaseAccount(address string, account state.UserAccountHandler) accountResponse {
	return accountResponse{
		Address:  address,
		Nonce:    account.GetNonce(),
		Balance:  account.GetBalance().String(),
		Code:     hex.EncodeToString(account.GetCode()),
		CodeHash: account.GetCodeHash(),
		RootHash: account.GetRootHash(),
	}
}
