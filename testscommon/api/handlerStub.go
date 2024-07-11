package api

import "net/http"

// HandlerStub -
type HandlerStub struct {
	ServeHTTPCalled func(writer http.ResponseWriter, request *http.Request)
}

// ServeHTTP -
func (stub *HandlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if stub.ServeHTTPCalled != nil {
		stub.ServeHTTPCalled(writer, request)
	}
}
