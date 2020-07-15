package block

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/gin-gonic/gin"
)

const (
	getBlockByNonceEndpoint = "/block/by-nonce/:nonce"
	getBlockByHashEndpoint  = "/block/by-hash/:hash"
)

// BlkService interface defines methods that can be used from `elrondFacade` context variable
type BlkService interface {
	GetBlockByHash(hash string) (*APIBlock, error)
	GetBlockByNonce(nonce uint64) (*APIBlock, error)
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
}

// APIBlock represents the structure for block that is returned by api routes
type APIBlock struct {
	Nonce      uint64   `form:"nonce" json:"nonce"`
	Round      uint64   `form:"round" json:"round"`
	Hash       string   `form:"hash" json:"hash"`
	Epoch      uint32   `form:"epoch" json:"epoch"`
	ShardID    uint32   `form:"shardID" json:"shardID"`
	NumTxs     uint32   `form:"numTxs" json:"numTxs"`
	MiniBlocks []string `form:"miniBlocks" json:"miniBlocks"`
}

// Routes defines block related routes
func Routes(routes *wrapper.RouterWrapper) {
	routes.RegisterHandler(http.MethodGet, "/by-nonce/:nonce", getBlockByNonce)
	routes.RegisterHandler(http.MethodGet, "/by-hash/:hash", getBlockByHash)
}

func getBlockByNonce(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(BlkService)
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

	endpointThrottler, ok := ef.GetThrottlerForEndpoint(getBlockByNonceEndpoint)
	if ok {
		if !endpointThrottler.CanProcess() {
			c.JSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: errors.ErrTooManyRequests.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
			return
		}

		endpointThrottler.StartProcessing()
		defer endpointThrottler.EndProcessing()
	}

	nonceStr := c.Param("nonce")
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if nonceStr == "" || err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	block, err := ef.GetBlockByNonce(nonce)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrGetBlock.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"block": block},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func getBlockByHash(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(BlkService)
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

	endpointThrottler, ok := ef.GetThrottlerForEndpoint(getBlockByHashEndpoint)
	if ok {
		if !endpointThrottler.CanProcess() {
			c.JSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: errors.ErrTooManyRequests.Error(),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
			return
		}

		endpointThrottler.StartProcessing()
		defer endpointThrottler.EndProcessing()
	}

	hash := c.Param("hash")
	if hash == "" {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	block, err := ef.GetBlockByHash(hash)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrGetBlock.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"block": block},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
