package block

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/gin-gonic/gin"
)

const (
	getBlockByNoncePath = "/by-nonce/:nonce"
	getBlockByHashPath  = "/by-hash/:hash"
)

var log = logger.GetOrCreate("api/block")

// BlockService interface defines methods that can be used from `elrondFacade` context variable
type BlockService interface {
	GetBlockByHash(hash string, withTxs bool) (*api.Block, error)
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
}

// Routes defines block related routes
func Routes(routes *wrapper.RouterWrapper) {
	routes.RegisterHandler(http.MethodGet, getBlockByNoncePath, getBlockByNonce)
	routes.RegisterHandler(http.MethodGet, getBlockByHashPath, getBlockByHash)
}

func getBlockByNonce(c *gin.Context) {
	ef, ok := getFacade(c)
	if !ok {
		return
	}

	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ef.GetBlockByNonce(nonce, withTxs)
	log.Debug(fmt.Sprintf("GetBlockByNonce took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

}

func getBlockByHash(c *gin.Context) {
	ef, ok := getFacade(c)
	if !ok {
		return
	}

	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ef.GetBlockByHash(hash, withTxs)
	log.Debug(fmt.Sprintf("GetBlockByHash took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func getQueryParamWithTxs(c *gin.Context) (bool, error) {
	withTxsStr := c.Request.URL.Query().Get("withTxs")
	if withTxsStr == "" {
		return false, nil
	}

	return strconv.ParseBool(withTxsStr)
}

func getQueryParamNonce(c *gin.Context) (uint64, error) {
	nonceStr := c.Param("nonce")
	if nonceStr == "" {
		return 0, errors.ErrInvalidBlockNonce
	}

	return strconv.ParseUint(nonceStr, 10, 64)
}

func getFacade(c *gin.Context) (BlockService, bool) {
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

	facade, ok := facadeObj.(BlockService)
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
