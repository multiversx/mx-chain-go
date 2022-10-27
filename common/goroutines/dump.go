package goroutines

import "runtime"

const oneHundredMBs = 1 << 20 * 100

// GetGoRoutines will try to get all go routines info from the running runtime into a string variable
// TODO move this to core package
func GetGoRoutines() string {
	buff := make([]byte, oneHundredMBs)
	n := runtime.Stack(buff, true)
	return string(buff[:n])
}
