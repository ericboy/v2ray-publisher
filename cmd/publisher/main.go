package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"

	publisher "github.com/ericboy/v2rayN-Publisher"
)

// BuildVersion metadata
var BuildVersion = "0.1.0"

// BuildGitHash metadata
var BuildGitHash = "(Not provided)"

// BuildTime metadata
var BuildTime = "(Not provided)"

func main() {
	// Define and parse command line parameters
	cfg := flag.String("config", "config.json", "Path to config file")
	certFile := flag.String("certFile", "", "TLS server cert file")
	keyFile := flag.String("keyFile", "", "TLS server key file")
	httpAddress := flag.String("address", "localhost:3000", "Address the HTTP Server listen to")
	flag.Parse()

	// Create publisher instance
	fmt.Println("v2rayN Publisher " + BuildVersion)
	fmt.Println("Build Time: " + BuildTime)
	fmt.Println("Commit Git hash: " + BuildGitHash)
	fmt.Printf("Runtime: %s (%s), %s\n\n", runtime.GOOS, runtime.GOARCH, runtime.Version())

	pub := publisher.NewPublisher()
	if err := pub.LoadConfigFile(*cfg); err != nil {
		return
	}

	// Create and start the HTTP service
	server := &http.Server{
		Addr:    *httpAddress,
		Handler: pub.Router(),
	}

	if *certFile != "" && *keyFile != "" {
		fmt.Println("[INFO] Publisher listen on: https://" + *httpAddress)
		server.ListenAndServeTLS(*certFile, *keyFile)
	} else {
		fmt.Printf("[WARN] Publisher running without HTTPS, this is not recommended. It may disclose your server information to unauthorized people, causing serious damage! Unless using reverse proxy like nginx with HTTPS, you SHOULD NOT do this at any time!\n")
		fmt.Println("[INFO] Publisher listen on: http://" + *httpAddress)
		server.ListenAndServe()
	}
}
