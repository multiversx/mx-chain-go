package hardfork

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/gin-gonic/gin"
)

const (
	execManualTrigger    = "executed, trigger is affecting only the current node"
	execBroadcastTrigger = "executed, trigger is affecting current node and will get broadcast to other peers"
	triggerPath          = "/trigger"
)

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	Trigger(epoch uint32, forced bool) error
	IsSelfTrigger() bool
	IsInterfaceNil() bool
}

// HarforkRequest represents the structure on which user input for triggering a hardfork will validate against
type HarforkRequest struct {
	Epoch  uint32 `form:"epoch" json:"epoch"`
	Forced bool   `form:"forced" json:"forced"`
}

// Routes defines node related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodPost, triggerPath, Trigger)
}

func getFacade(c *gin.Context) (FacadeHandler, bool) {
	facadeObj, ok := c.Get("facade")
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNilAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	facade, ok := facadeObj.(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	return facade, true
}

// Trigger will receive a trigger request from the client and propagate it for processing
func Trigger(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	var hr = HarforkRequest{}
	err := c.ShouldBindJSON(&hr)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	err = facade.Trigger(hr.Epoch, hr.Forced)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	status := execManualTrigger
	if facade.IsSelfTrigger() {
		status = execBroadcastTrigger
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"status": status},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
