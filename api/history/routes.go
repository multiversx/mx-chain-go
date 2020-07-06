package history

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/gin-gonic/gin"
)

// HistoryTransaction is structure that holds information about a history transaction
type HistoryTransaction struct {
	Round      uint64 `json:"round"`
	SndShard   uint32 `json:"sndShardID"`
	RcvShard   uint32 `json:"rcvShardID"`
	BlockNonce uint64 `json:"blockNonce"`
	MBHash     string `json:"miniblockHash"`
	BlockHash  string `json:"blockHash"`
	Status     string `json:"status"`
}

// HistoryService -
type HistoryService interface {
	GetHistoryTransaction(hash string) (*HistoryTransaction, error)
}

// Routes defines history related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, "/transaction/:txhash", GetHistoryTransaction)
}

// GetHistoryTransaction returns transaction details for a given txhash
func GetHistoryTransaction(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(HistoryService)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyTxHash.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	tx, err := ef.GetHistoryTransaction(txhash)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrGetTransaction.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"transaction": tx},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
