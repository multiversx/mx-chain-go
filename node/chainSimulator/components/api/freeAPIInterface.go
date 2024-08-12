package api

import (
	"fmt"
	"net"
)

type freePortAPIConfigurator struct {
	restAPIInterface string
}

// NewFreePortAPIConfigurator will create a new instance of freePortAPIConfigurator
func NewFreePortAPIConfigurator(restAPIInterface string) *freePortAPIConfigurator {
	return &freePortAPIConfigurator{
		restAPIInterface: restAPIInterface,
	}
}

// RestApiInterface will return the rest api interface with a free port
func (f *freePortAPIConfigurator) RestApiInterface(_ uint32) string {
	return fmt.Sprintf("%s:%d", f.restAPIInterface, getFreePort())
}

func getFreePort() int {
	// Listen on port 0 to get a free port
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = l.Close()
	}()

	// Get the port number that was assigned
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port
}
