package main

import (
	"log"
	"net/http"

	"github.com/Stratoscale/swagger/example/auth"
	"github.com/Stratoscale/swagger/example/internal"
	"github.com/Stratoscale/swagger/example/restapi"
)

func main() {
	// Initiate business logic implementers.
	// This is the main function, so here the implementers' dependencies can be
	// injected, such as database, parameters from environment variables, or different
	// clients for different APIs.
	p := internal.Pet{}
	s := internal.Store{}

	// Initiate the http handler, with the objects that are implementing the business logic.
	h, err := restapi.Handler(restapi.Config{
		PetAPI:     &p,
		StoreAPI:   &s,
		AuthToken:  auth.Token,
		Authorizer: auth.Request,
		Logger:     log.Printf,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Run the standard http server
	log.Fatal(http.ListenAndServe(":8080", h))
}
