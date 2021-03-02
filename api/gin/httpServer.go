package gin

import (
	"context"
	"net/http"
	"time"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
)

type httpServer struct {
	server *http.Server
}

// NewHttpServer returns a new instance of httpServer
func NewHttpServer(server *http.Server) (*httpServer, error) {
	if server == nil {
		return nil, apiErrors.ErrNilHttpServer
	}

	return &httpServer{
		server: server,
	}, nil
}

// Start will handle the starting of the gin web server. This call is blocking and it should be
// called on a go routine (different than the main one)
func (h *httpServer) Start() {
	err := h.server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			log.Error("could not start webserver",
				"error", err.Error(),
			)
		} else {
			log.Debug("ListenAndServe - webserver closed")
		}
	}
}

// Stop will handle the stopping of the gin web server
func (h *httpServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return h.server.Shutdown(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *httpServer) IsInterfaceNil() bool {
	return h == nil
}
