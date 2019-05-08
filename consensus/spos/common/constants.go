package common

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
)

var log = logger.DefaultLogger()

const RoundTimeDuration = time.Duration(100 * time.Millisecond)
