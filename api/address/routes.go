package address

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	GetBalance(address string) (*big.Int, error)
	GetAccount(address string) (*state.Account, error)
}

type accountResponse struct {
	Address  string `json:"address"`
	Nonce    uint64 `json:"nonce"`
	Balance  string `json:"balance"`
	CodeHash []byte `json:"codeHash"`
	RootHash []byte `json:"rootHash"`
}

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/:address", GetAccount)
	router.GET("/:address/balance", GetBalance)
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

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func accountResponseFromBaseAccount(address string, account *state.Account) accountResponse {
	return accountResponse{
		Address:  address,
		Nonce:    account.Nonce,
		Balance:  account.Balance.String(),
		CodeHash: account.CodeHash,
		RootHash: account.RootHash,
	}
}
