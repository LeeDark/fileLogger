package main

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

const REST_PORT = 5100

// RestService provides HTTP service.
type RestService struct {
	address string
	ln      net.Listener
}

// New returns an uninitialized HTTP service.
func NewRestService(address string) *RestService {
	return &RestService{
		address: address,
	}
}

// Start starts the service.
func (s *RestService) Start() error {
	// Get the mux router object
	router := mux.NewRouter().StrictSlash(false)

	// curl -X GET localhost:5100/
	router.HandleFunc("/", Home)

	// /api/files
	// curl -F "filename=@/home/sergey/test.txt" -X POST localhost:5100/api/files
	//router.HandleFunc("/api/files", PostFile).Methods("POST")
	// curl -X GET localhost:5100/api/files/{test.txt} --output test.txt
	//router.HandleFunc("/api/files", GetFile).Methods("GET")

	// Create a negroni instance
	n := negroni.Classic()
	corsHandler := cors.AllowAll().Handler(router)
	n.UseHandler(corsHandler)

	server := http.Server{
		Handler: n,
	}

	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.ln = ln

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			Info.Printf("HTTP serve: %s", err)
		}
		// FIXME
		//shutdown <- 1
	}()

	return nil
}

// Close closes the service.
func (s *RestService) Close() {
	Info.Println("rest closing")
	s.ln.Close()
	return
}
