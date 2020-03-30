package hardfork

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/gin-gonic/gin"
)

const notExecuted = "not executed, trigger is not requested"
const execManualTrigger = "executed, trigger is affecting only the current node"
const execBroadcastTrigger = "executed, trigger is affecting current node and will get broadcast to other peers"

// TriggerHardforkHandler interface defines methods that can be used from `elrondFacade` context variable
type TriggerHardforkHandler interface {
	Trigger() error
	IsSelfTrigger() bool
}

// TriggerHardforkRequest represents the structure that is posted by the client, requesting a trigger
type TriggerHardforkRequest struct {
	Triggered bool
}

// Routes defines node related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/trigger", Trigger)
}

// Trigger will receive a trigger request from the client and propagate it for processing
func Trigger(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TriggerHardforkHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gthr = TriggerHardforkRequest{}
	err := c.ShouldBindJSON(&gthr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	if !gthr.Triggered {
		c.JSON(http.StatusOK, gin.H{"status": notExecuted})
		return
	}

	err = ef.Trigger()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := execManualTrigger
	if ef.IsSelfTrigger() {
		status = execBroadcastTrigger
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}
