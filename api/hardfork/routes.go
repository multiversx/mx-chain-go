package hardfork

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/gin-gonic/gin"
)

const notExecuted = "not executed"
const execManualTrigger = "executed, manual self"
const execBroadcastTrigger = "executed, broadcast trigger"

// HardforkHandler interface defines methods that can be used from `elrondFacade` context variable
type HardforkHandler interface {
	Trigger() error
	IsSelfTrigger() bool
}

// TriggerHardforkRequest represents the structure that is posted by the client, requesting a trigger
type TriggerHardforkRequest struct {
	Trigger bool
}

// Routes defines node related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/trigger", Trigger)
}

// Trigger will receive a trigger request from the client and propagate it for processing
func Trigger(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(HardforkHandler)
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

	if !gthr.Trigger {
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
