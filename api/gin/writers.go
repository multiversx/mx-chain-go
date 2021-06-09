package gin

import "bytes"

type ginWriter struct {
}

// Write will output the message using elrond-go-logger's logger
func (gv *ginWriter) Write(p []byte) (n int, err error) {
	trimmed := bytes.TrimSpace(p)
	log.Trace("gin server", "message", string(trimmed))

	return len(p), nil
}

type ginErrorWriter struct {
}

// Write will output the error using elrond-go-logger's logger
func (gev *ginErrorWriter) Write(p []byte) (n int, err error) {
	trimmed := bytes.TrimSpace(p)
	log.Trace("gin server", "error", string(trimmed))

	return len(p), nil
}
