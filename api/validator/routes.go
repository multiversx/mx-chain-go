package validator

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/gin-gonic/gin"
)

// ValidatorsStatisticsApiHandler interface defines methods that can be used from `elrondFacade` context variable
type ValidatorsStatisticsApiHandler interface {
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	IsInterfaceNil() bool
}

// Routes defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.GET("/statistics", Statistics)
}

// Statistics will return the validation statistics for all validators
func Statistics(c *gin.Context) {
	ef, ok := c.MustGet("elrondFacade").(ValidatorsStatisticsApiHandler)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	valStats, err := ef.ValidatorStatisticsApi()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": valStats})
}
