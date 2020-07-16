package middleware

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/gin-gonic/gin"
)

type throttlerGetter interface {
	GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool)
}

// CreateEndpointThrottler will create a middleware-type of handler to be used in conjunction with special
// REST API end points that need to be better protected
func CreateEndpointThrottler(throttlerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tgObj, ok := c.Get("facade")
		if !ok {
			c.JSON(
				http.StatusInternalServerError,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: errors.ErrNilAppContext.Error(),
					Code:  shared.ReturnCodeInternalError,
				},
			)
			return
		}

		tg, ok := tgObj.(throttlerGetter)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: errors.ErrInvalidAppContext.Error(),
					Code:  shared.ReturnCodeInternalError,
				},
			)
			return
		}

		endpointThrottler, ok := tg.GetThrottlerForEndpoint(throttlerName)
		if !ok {
			c.Next()
			return
		}

		if !endpointThrottler.CanProcess() {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				shared.GenericAPIResponse{
					Data:  nil,
					Error: fmt.Sprintf("%s for endpoint %s", errors.ErrTooManyRequests.Error(), throttlerName),
					Code:  shared.ReturnCodeSystemBusy,
				},
			)
			return
		}

		endpointThrottler.StartProcessing()
		defer endpointThrottler.EndProcessing()

		c.Next()
	}
}
