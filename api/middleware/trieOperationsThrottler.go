package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

type trieOperationsThrottler struct {
	slots chan struct{}
}

// NewTrieOperationsThrottler holds admission through logging and response flushing.
func NewTrieOperationsThrottler(cfg config.TrieOperationsConfig) *trieOperationsThrottler {
	return &trieOperationsThrottler{slots: make(chan struct{}, cfg.WithDefaults().MaxConcurrentRequests)}
}

func (t *trieOperationsThrottler) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isTrieListing(c) {
			c.Next()
			return
		}
		select {
		case t.slots <- struct{}{}:
			defer func() { <-t.slots }()
		default:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, shared.GenericAPIResponse{
				Error: common.ErrTrieOperationsBusy.Error(), Code: shared.ReturnCodeSystemBusy,
			})
			return
		}
		c.Next()
		_ = http.NewResponseController(c.Writer).Flush()
	}
}

func isTrieListing(c *gin.Context) bool {
	if c.Request.Method == http.MethodPost {
		return c.FullPath() == "/address/iterate-keys"
	}
	if c.Request.Method != http.MethodGet {
		return false
	}
	switch c.FullPath() {
	case "/address/:address":
		withKeys, _ := strconv.ParseBool(c.Query("withKeys"))
		return withKeys
	case "/address/:address/keys", "/address/:address/esdt", "/address/:address/esdts/roles",
		"/address/:address/esdts-with-role/:role", "/address/:address/registered-nfts",
		"/network/esdts", "/network/esdt/fungible-tokens", "/network/esdt/semi-fungible-tokens",
		"/network/esdt/non-fungible-tokens":
		return true
	}
	return false
}

func (t *trieOperationsThrottler) IsInterfaceNil() bool { return t == nil }
