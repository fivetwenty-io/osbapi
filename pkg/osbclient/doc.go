// Package osbclient provides the platform-side HTTP client for calling
// Open Service Broker API v2.17 compliant service brokers.
//
// Platforms use [New] or [NewWithBasicAuth] to create a client, then call
// methods corresponding to OSB API operations.
//
// Example:
//
//	client, err := osbclient.NewWithBasicAuth("https://broker.example.com", "user", "pass")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	catalog, err := client.GetCatalog(ctx)
package osbclient
