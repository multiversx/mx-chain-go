package hardfork

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/gin-gonic/gin"
)

const execManualTrigger = "executed, trigger is affecting only the current node"
const execBroadcastTrigger = "executed, trigger is affecting current node and will get broadcast to other peers"

// TriggerHardforkHandler interface defines methods that can be used from `elrondFacade` context variable
type TriggerHardforkHandler interface {
	Trigger() error
	IsSelfTrigger() bool
}

// Routes defines node related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodPost, "/trigger", Trigger)
}

// Trigger will receive a trigger request from the client and propagate it for processing
func Trigger(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(TriggerHardforkHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	err := ef.Trigger()
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
