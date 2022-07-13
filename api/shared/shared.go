package shared

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MiddlewarePosition is the type that specifies the position of a middleware relative to the base endpoint handler
type MiddlewarePosition bool

const (
	// Before indicates that the middleware should be used before the base endpoint handler
	Before MiddlewarePosition = true

	// After indicates that the middleware should be used after the base endpoint handler
	After MiddlewarePosition = false
)

// AdditionalMiddleware holds the data needed for adding a middleware to an API endpoint
type AdditionalMiddleware struct {
	Middleware gin.HandlerFunc
	Position   MiddlewarePosition
}

// EndpointHandlerData holds the items needed for creating a new gin HTTP endpoint
type EndpointHandlerData struct {
	Path                  string
	Method                string
	Handler               gin.HandlerFunc
	AdditionalMiddlewares []AdditionalMiddleware
}

// GenericAPIResponse defines the structure of all responses on API endpoints
type GenericAPIResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  ReturnCode  `json:"code"`
}

// ReturnCode defines the type defines to identify return codes
type ReturnCode string

// ReturnCodeSuccess defines a successful request
const ReturnCodeSuccess ReturnCode = "successful"

// ReturnCodeInternalError defines a request which hasn't been executed successfully due to an internal error
const ReturnCodeInternalError ReturnCode = "internal_issue"

// ReturnCodeRequestError defines a request which hasn't been executed successfully due to a bad request received
const ReturnCodeRequestError ReturnCode = "bad_request"

// ReturnCodeSystemBusy defines a request which hasn't been executed successfully due to too many requests
const ReturnCodeSystemBusy ReturnCode = "system_busy"

// RespondWith will respond with the generic API response
func RespondWith(c *gin.Context, status int, dataField interface{}, errMessage string, code ReturnCode) {
	c.JSON(
		status,
		GenericAPIResponse{
			Data:  dataField,
			Error: errMessage,
			Code:  code,
		},
	)
}

// RespondWithValidationError should be called when the request cannot be satisfied due to a (request) validation error
func RespondWithValidationError(c *gin.Context, err error, innerErr error) {
	errMessage := fmt.Sprintf("%s: %s", err.Error(), innerErr.Error())

	RespondWith(
		c,
		http.StatusBadRequest,
		nil,
		errMessage,
		ReturnCodeRequestError,
	)
}

// RespondWithInternalError should be called when the request cannot be satisfied due to an internal error
func RespondWithInternalError(c *gin.Context, err error, innerErr error) {
	errMessage := fmt.Sprintf("%s: %s", err.Error(), innerErr.Error())

	RespondWith(
		c,
		http.StatusInternalServerError,
		nil,
		errMessage,
		ReturnCodeInternalError,
	)
}

// RespondWithSuccess should be called when the request can be satisfied
func RespondWithSuccess(c *gin.Context, data interface{}) {
	RespondWith(
		c,
		http.StatusOK,
		data,
		"",
		ReturnCodeSuccess,
	)
}
