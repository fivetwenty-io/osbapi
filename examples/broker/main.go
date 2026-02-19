package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fivetwenty-io/osbapi/v2/pkg/broker"
)

func main() {
	b := NewInMemoryBroker()

	handler := broker.NewHandler(b,
		broker.WithBasicAuth("broker", "secret"),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting OSB API broker on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
