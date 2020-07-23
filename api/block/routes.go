package block

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/gin-gonic/gin"
)

const (
	getBlockByNoncePath = "/by-nonce/:nonce"
	getBlockByHashPath  = "/by-hash/:hash"
)

// BlockService interface defines methods that can be used from `elrondFacade` context variable
type BlockService interface {
	GetBlockByHash(hash string, withTxs bool) (*APIBlock, error)
	GetBlockByNonce(nonce uint64, withTxs bool) (*APIBlock, error)
}

// APIBlock represents the structure for block that is returned by api routes
type APIBlock struct {
	Nonce                uint64          `form:"nonce" json:"nonce"`
	Round                uint64          `form:"round" json:"round"`
	Hash                 string          `form:"hash" json:"hash"`
	Epoch                uint32          `form:"epoch" json:"epoch"`
	ShardID              uint32          `form:"shardID" json:"shardID"`
	NumTxs               uint32          `form:"numTxs" json:"numTxs"`
	NotarizedBlockHashes []string        `form:"notarizedBlockHashes" json:"notarizedBlockHashes,omitempty"`
	MiniBlocks           []*APIMiniBlock `form:"miniBlocks" json:"miniBlocks,omitempty"`
}

// APIMiniBlock represents the structure for a miniblock
type APIMiniBlock struct {
	Hash               string                              `form:"hash" json:"hash"`
	Type               string                              `form:"type" json:"type"`
	SourceShardID      uint32                              `form:"sourceShardID" json:"sourceShardID"`
	DestinationShardID uint32                              `form:"destinationShardID" json:"destinationShardID"`
	Transactions       []*transaction.ApiTransactionResult `form:"transactions" json:"transactions,omitempty"`
}

// Routes defines block related routes
func Routes(routes *wrapper.RouterWrapper) {
	routes.RegisterHandler(http.MethodGet, getBlockByNoncePath, getBlockByNonce)
	routes.RegisterHandler(http.MethodGet, getBlockByHashPath, getBlockByHash)
}

func getBlockByNonce(c *gin.Context) {
	ef, ok := c.MustGet("facade").(BlockService)
	if !ok {
		shared.RespondWithInvalidAppContext(c)
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

	block, err := ef.GetBlockByNonce(nonce, withTxs)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			errors.ErrGetBlock.Error(),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

}

func getBlockByHash(c *gin.Context) {
	ef, ok := c.MustGet("facade").(BlockService)
	if !ok {
		shared.RespondWithInvalidAppContext(c)
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

	block, err := ef.GetBlockByHash(hash, withTxs)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			errors.ErrGetBlock.Error(),
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
