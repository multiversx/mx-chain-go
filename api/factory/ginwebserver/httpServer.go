package ginwebserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type httpServer struct {
	server *http.Server
}

// NewHttpServer returns a new instance of httpServer
func NewHttpServer(server *http.Server) (*httpServer, error) {
	if server == nil {
		return nil, errors.New("nil http server")
	}

	return &httpServer{
		server: server,
	}, nil
}

// Start will handle the starting of the gin web server
func (g *httpServer) Start() {
	go func(server *http.Server) {
		err := server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				log.Error("could not start webserver",
					"error", err.Error(),
				)
			} else {
				log.Debug("ListenAndServe - webserver closed")
			}
		}
	}(g.server)
}

// Stop will handle the stopping of the gin web server
func (g *httpServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return g.server.Shutdown(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *httpServer) IsInterfaceNil() bool {
	return g == nil
}
