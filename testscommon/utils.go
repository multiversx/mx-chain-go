package testscommon

import (
	"fmt"
	"time"
)

func AddTimestampSuffix(tag string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s", tag, timestamp)
}

func panicIfError(message string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", message, err))
	}
}
