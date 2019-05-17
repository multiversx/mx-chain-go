package block

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	RecentNotarizedBlocks(maxShardHeadersNum int) ([]*external.BlockHeader, error)
	RetrieveShardBlock(blockHash []byte) (*external.ShardBlockInfo, error)
}

type blockResponse struct {
	Nonce         uint64   `json:"nonce"`
	ShardID       uint32   `json:"shardId"`
	Hash          string   `json:"hash"`
	Proposer      string   `json:"proposer"`
	Validators    []string `json:"validators"`
	PubKeyBitmap  string   `json:"pubKeyBitmap"`
	Size          int64    `json:"size"`
	Timestamp     uint64   `json:"timestamp"`
	TxCount       uint32   `json:"txCount"`
	StateRootHash string   `json:"stateRootHash"`
	PrevHash      string   `json:"prevHash"`
}

func formattedRecentBlocks(headers []*external.BlockHeader) []blockResponse {
	frb := make([]blockResponse, len(headers))

	for index, header := range headers {
		frb[index] = formatShardHeader(header)
	}

	return frb
}

func formatShardHeader(header *external.BlockHeader) blockResponse {
	return blockResponse{
		Nonce:    header.Nonce,
		ShardID:  header.ShardID,
		Hash:     hex.EncodeToString(header.Hash),
		Proposer: hex.EncodeToString(header.ProposerPubKey),
		// TODO: Add all validators
		Validators:    []string{hex.EncodeToString(header.ProposerPubKey)},
		PubKeyBitmap:  hex.EncodeToString(header.PubKeysBitmap),
		Size:          header.BlockSize,
		Timestamp:     header.Timestamp,
		TxCount:       header.TxCount,
		StateRootHash: hex.EncodeToString(header.StateRootHash),
		PrevHash:      hex.EncodeToString(header.PrevHash),
	}
}

// Routes defines block related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/:block", Block)
}

// RoutesForBlockLists defines routes related to lists of blocks. Used sepparatly so
//  it will not confloct with the wildcard for block details route
func RoutesForBlockLists(router *gin.RouterGroup) {
	router.GET("/recent", RecentBlocks)
}

// Block returns a single blockResponse object containing information
//  about the requested block associated with the provided hash
func Block(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	blockHashStringified := c.Param("block")
	//Change here if representation (request) of a block hash changes from hex
	blockHash, err := hex.DecodeString(blockHashStringified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	headerInfo, err := ef.RetrieveShardBlock([]byte(blockHash))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": formatShardHeader(&headerInfo.BlockHeader)})
}

// RecentBlocks returns a list of blockResponse objects containing most
//  recent blocks from each shard
func RecentBlocks(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	recentBlocks, err := ef.RecentNotarizedBlocks(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blocks": formattedRecentBlocks(recentBlocks)})
}
