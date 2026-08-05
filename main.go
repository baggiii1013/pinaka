package main

import (
	"os"

	"github.com/charmbracelet/log"
	"pinakatype.sh/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "2222"
	}

	log.SetLevel(log.DebugLevel)
	server.Start(port)
}