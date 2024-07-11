package gin

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
)

const closedNetworkConnection = "use of closed network connection"

type httpServer struct {
	listener net.Listener
	handler  http.Handler
}

// NewHttpServer returns a new instance of httpServer
func NewHttpServer(address string, network string, handler http.Handler) (*httpServer, error) {
	if check.IfNilReflect(handler) {
		return nil, apiErrors.ErrNilHTTPHandler
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &httpServer{
		listener: listener,
		handler:  handler,
	}, nil
}

// Start will handle the starting of the gin web server. This call is blocking, and it should be
// called on a go routine (different from the main one)
func (h *httpServer) Start() {
	err := http.Serve(h.listener, h.handler)
	if err == nil {
		return
	}

	if errors.Is(err, http.ErrServerClosed) || strings.Contains(err.Error(), closedNetworkConnection) {
		log.Debug("ListenAndServe - webserver closed")
		return
	}

	log.Error("could not start webserver",
		"error", err.Error(),
	)
}

// Close will handle the stopping of the gin web server
func (h *httpServer) Close() error {
	return h.listener.Close()
}

// Port returns the port number this server was bound to
// TODO: use this method by the facade, to report the bound port
func (h *httpServer) Port() int {
	return h.listener.Addr().(*net.TCPAddr).Port
}

// Address returns the actual address this server was bound to
// TODO: use this method by the facade, to report the bound address
func (h *httpServer) Address() string {
	return h.listener.Addr().String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *httpServer) IsInterfaceNil() bool {
	return h == nil
}
