package shared

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
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
func RespondWith(c *gin.Context, status int, dataField interface{}, error string, code ReturnCode) {
	c.JSON(
		status,
		GenericAPIResponse{
			Data:  dataField,
			Error: error,
			Code:  code,
		},
	)
}

// RespondWithInvalidAppContext will be called when the application's context is invalid
func RespondWithInvalidAppContext(c *gin.Context) {
	RespondWith(
		c,
		http.StatusInternalServerError,
		nil,
		errors.ErrInvalidAppContext.Error(),
		ReturnCodeInternalError,
	)
}

// RespondWithValidationError will be called when the application's context is invalid
func RespondWithValidationError(c *gin.Context, error string) {
	RespondWith(
		c,
		http.StatusBadRequest,
		nil,
		error,
		ReturnCodeRequestError,
	)
}
