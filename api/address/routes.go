package address

import (
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
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
//  about the account corelated with provided address
func GetAccount(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}

	addr := c.Param("address")
	acc, err := ef.GetAccount(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get requested account: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": accountResponseFromBaseAccount(addr, acc)})
}

//GetBalance returns the balance for the address parameter
func GetBalance(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid app context"})
		return
	}
	addr := c.Param("address")

	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Get balance error: Address was empty"})
		return
	}

	balance, err := ef.GetBalance(addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Get balance error: " + err.Error()})
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
