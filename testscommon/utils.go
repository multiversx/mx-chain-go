package testscommon

import (
	"fmt"
	"time"
)

// HashSize holds the size of a typical hash used by the protocol
const HashSize = 32

// AddTimestampSuffix -
func AddTimestampSuffix(tag string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s", tag, timestamp)
}
