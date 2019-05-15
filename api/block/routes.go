package block

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler interface defines methods that can be used from `elrondFacade` context variable
type Handler interface {
	RecentNotarizedBlocks() interface{}
}

type blockResponse struct {
	Nonce     uint64 `json:"nonce"`
	ShardID   uint32 `json:"shardId"`
	Hash      string `json:"hash"`
	Proposer  string `json:"proposer"`
	Size      string `json:"size"`
	Timestamp int64 `json:"timestamp"`
}

type recentBlocksResponse struct {
	Blocks []blockResponse `json:"blocks"`
}

func buildDummyBlock() blockResponse {
	return blockResponse{
		1, 1, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"32kb", 23,
	}
}

func buildDummyRecentBlocks() recentBlocksResponse {
	recentBlocks := make([]blockResponse, 0)
	for i := 0; i < 10; i++ {
		recentBlocks = append(recentBlocks, buildDummyBlock())
	}
	return recentBlocksResponse{recentBlocks}
}

// Routes defines block related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/:block", Block)
	router.GET("/recent-blocks", RecentBlocks)
}

// Block returns a single blockResponse object containing information
//  about the requested block associated with the provided hash
func Block(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"block": buildDummyBlock()})
}

// RecentBlocks returns a list of blockResponse objects containing most
//  recent blocks from each shard
func RecentBlocks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"blocks": buildDummyRecentBlocks().Blocks})
}
