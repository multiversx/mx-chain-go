package api

import (
	"fmt"
	"net"
	"sync"
)

type freePortAPIConfigurator struct {
	restAPIInterface string
	mut              sync.Mutex
	allocatedPorts   map[int]struct{}
	portProvider     func() int
}

// NewFreePortAPIConfigurator will create a new instance of freePortAPIConfigurator
func NewFreePortAPIConfigurator(restAPIInterface string, reservedPorts ...int) *freePortAPIConfigurator {
	configurator := &freePortAPIConfigurator{
		restAPIInterface: restAPIInterface,
		allocatedPorts:   make(map[int]struct{}, len(reservedPorts)),
		portProvider:     getFreePort,
	}

	for _, port := range reservedPorts {
		if port > 0 {
			configurator.allocatedPorts[port] = struct{}{}
		}
	}

	return configurator
}

// RestApiInterface will return the rest api interface with a free port
func (f *freePortAPIConfigurator) RestApiInterface(_ uint32) string {
	f.mut.Lock()
	defer f.mut.Unlock()

	for {
		port := f.portProvider()
		if _, alreadyAllocated := f.allocatedPorts[port]; alreadyAllocated {
			continue
		}

		f.allocatedPorts[port] = struct{}{}
		return fmt.Sprintf("%s:%d", f.restAPIInterface, port)
	}
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
