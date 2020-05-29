package address

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `elrondFacade` context variable
type FacadeHandler interface {
	GetBalance(address string) (*big.Int, error)
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
}

// GetAccount returns an accountResponse containing information
//  about the account correlated with provided address
func GetAccount(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}

	addr := c.Param("address")
	acc, err := ef.GetAccount(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrCouldNotGetAccount.Error(), err.Error()),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		core.GenericAPIResponse{
			Data:  gin.H{"account": accountResponseFromBaseAccount(addr, acc)},
			Error: "",
			Code:  string(core.ReturnCodeSuccess),
		},
	)
}

// GetBalance returns the balance for the address parameter
func GetBalance(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}
	addr := c.Param("address")
	if addr == "" {
		c.JSON(
			http.StatusBadRequest,
			core.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), errors.ErrEmptyAddress.Error()),
				Code:  string(core.ReturnCodeRequestErrror),
			},
		)
		return
	}

	balance, err := ef.GetBalance(addr)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			core.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetBalance.Error(), err.Error()),
				Code:  string(core.ReturnCodeInternalError),
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		core.GenericAPIResponse{
			Data:  gin.H{"balance": balance.String()},
			Error: "",
			Code:  string(core.ReturnCodeSuccess),
		},
	)
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
