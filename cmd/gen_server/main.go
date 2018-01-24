// Package main has a file server that exposes a directory and listens on a port
// that are passed in flags.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jwowillo/gen/server"
)

// main runs a file server that exposes a dierctory and listens on a port that
// are passed in flags.
func main() {
	log.Printf("listening on %s and serving files from %s\n", port, dir)
	http.ListenAndServe(port, server.Handler(dir))
}

var (
	// port to listen on.
	port string
	// dir to expose.
	dir string
)

// init parses flags into variables.
func init() {
	flag.StringVar(&port, "port", "", "port to serve from")
	flag.StringVar(&dir, "directory", "", "directory with static files")
	flag.Parse()
	if port == "" {
		log.Fatal("must pass port to serve from")
	}
	if dir == "" {
		log.Fatal("must pass directory with static files")
	}
}
