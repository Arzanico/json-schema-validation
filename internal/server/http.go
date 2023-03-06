package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func NewHttpServer(addr string) *http.Server {
	httpsrv := newHttpServer()
	r := mux.NewRouter()
	r.HandleFunc("/validate", httpsrv.Validate).Methods(http.MethodPost)
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type httpServer struct {
	Payload *PayloadValidationRequest
}

func newHttpServer() *httpServer {
	return &httpServer{
		Payload: NewPayloadValidationRequest(),
	}
}

func (s *httpServer) Validate( (w http.ResponseWriter, r *http.Request) {
	fmt.Println("Validation")
}
