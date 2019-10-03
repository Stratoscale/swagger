package main

import (
	"log"
	"net/http"

	"github.com/Stratoscale/swagger/testdata/auth"
	"github.com/Stratoscale/swagger/testdata/internal/pet"
	"github.com/Stratoscale/swagger/testdata/internal/store"
	"github.com/Stratoscale/swagger/testdata/restapi"
)

func main() {
	// Initiate business logic implementers.
	// This is the main function, so here the implementers' dependencies can be
	// injected, such as database, parameters from environment variables, or different
	// clients for different APIs.
	p := pet.New()
	s := store.Store{}

	// Initiate the http handler, with the objects that are implementing the business logic.
	h, err := restapi.Handler(restapi.Config{
		PetAPI:     p,
		StoreAPI:   &s,
		AuthToken:  auth.Token,
		Authorizer: auth.Request,
		Logger:     log.Printf,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting to serve, access server on http://localhost:8080")

	// Run the standard http server
	log.Fatal(http.ListenAndServe(":8080", h))
}
