package immunitycache

import (
	"fmt"
)

var errInvalidConfig = fmt.Errorf("invalid cache config")
var errFailedEviction = fmt.Errorf("failed eviction")
