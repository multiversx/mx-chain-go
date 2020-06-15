package testscommon

import (
	"fmt"
)

func panicIfError(message string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", message, err))
	}
}
