package block

import (
	"encoding/hex"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	RecentNotarizedBlocks(maxShardHeadersNum int) ([]external.RecentBlock, error)
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

const recentBlocksCount = 20

func buildDummyBlock() blockResponse {
	return blockResponse{
		1, 1, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		[]string{"0x1234567", "0x123242342"}, "0x1234", 32, 23, 15000,
		"0x883284732784278", "0x32184232364274",
	}
}

func formattedRecentBlocks(rb []external.RecentBlock) []blockResponse {
	frb := make([]blockResponse, len(rb))

	for index, block := range rb {
		frb[index] = blockResponse{
			Nonce:    block.Nonce,
			ShardID:  block.ShardID,
			Hash:     hex.EncodeToString(block.Hash),
			Proposer: hex.EncodeToString(block.ProposerPubKey),
			// TODO: Add all validators
			Validators:    []string{hex.EncodeToString(block.ProposerPubKey)},
			PubKeyBitmap:  hex.EncodeToString(block.PubKeysBitmap),
			Size:          block.BlockSize,
			Timestamp:     block.Timestamp,
			TxCount:       block.TxCount,
			StateRootHash: hex.EncodeToString(block.StateRootHash),
			PrevHash:      hex.EncodeToString(block.PrevHash),
		}
	}

	return frb
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
	c.JSON(http.StatusOK, gin.H{"block": buildDummyBlock()})
}

// RecentBlocks returns a list of blockResponse objects containing most
//  recent blocks from each shard
func RecentBlocks(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(Handler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	recentBlocks, err := ef.RecentNotarizedBlocks(recentBlocksCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blocks": formattedRecentBlocks(recentBlocks)})
}
