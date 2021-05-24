package shared

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
func RespondWith(c *gin.Context, status int, dataField interface{}, err string, code ReturnCode) {
	c.JSON(
		status,
		GenericAPIResponse{
			Data:  dataField,
			Error: err,
			Code:  code,
		},
	)
}

// RespondWithValidationError will be called when the application's context is invalid
func RespondWithValidationError(c *gin.Context, err string) {
	RespondWith(
		c,
		http.StatusBadRequest,
		nil,
		err,
		ReturnCodeRequestError,
	)
}
