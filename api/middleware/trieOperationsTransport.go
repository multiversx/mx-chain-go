package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/config"
)

type trieOperationsTransport struct {
	timeout      time.Duration
	reserve      time.Duration
	maxBodyBytes int64
}

// NewTrieOperationsTransport also bounds responses rejected by downstream limiters. The request context expires
// one reserve before the socket deadlines, so the handler can still write its response; a reserve that is not
// smaller than the timeout is ignored.
func NewTrieOperationsTransport(cfg config.TrieOperationsConfig) *trieOperationsTransport {
	cfg = cfg.WithDefaults()
	timeout := time.Duration(cfg.HTTPTimeoutMilliseconds) * time.Millisecond
	reserve := time.Duration(cfg.DeadlineReserveMilliseconds) * time.Millisecond
	if reserve >= timeout {
		log.Warn("trie operations deadline reserve is not smaller than the HTTP timeout, ignoring it",
			"reserve", reserve, "timeout", timeout)
		reserve = 0
	}
	return &trieOperationsTransport{
		timeout:      timeout,
		reserve:      reserve,
		maxBodyBytes: int64(cfg.MaxRequestBodyBytes),
	}
}

func (t *trieOperationsTransport) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isTrieOperation(c) {
			c.Next()
			return
		}
		controller := http.NewResponseController(c.Writer)
		deadline := time.Now().Add(t.timeout)
		// Keep deadlines active through body handling, logging and the final flush.
		defer func() {
			if time.Now().Before(deadline) {
				_ = controller.SetReadDeadline(time.Time{})
				_ = controller.SetWriteDeadline(time.Time{})
			}
		}()
		if err := controller.SetReadDeadline(deadline); err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		if err := controller.SetWriteDeadline(deadline); err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithDeadline(c.Request.Context(), deadline.Add(-t.reserve))
		defer cancel()
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, t.maxBodyBytes)
		c.Request = c.Request.WithContext(ctx)
		if c.Request.ContentLength > t.maxBodyBytes {
			c.Header("Connection", "close")
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			_ = controller.Flush()
			return
		}
		c.Next()
		_ = controller.Flush()
	}
}

func (t *trieOperationsTransport) IsInterfaceNil() bool { return t == nil }
