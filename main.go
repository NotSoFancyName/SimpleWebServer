package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NotSoFancyName/SimpleWebServer/server"
)

var port = flag.Int("p", 8081, "listen port number")

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	errs := make(chan error)
	stop := make(chan struct{})
	
	go server.NewServer(stop, *port).Run(errs)
	select {
	case err := <-errs:
		log.Printf("Server stopped due to error: %v", err)
	case <-done:
		stop <- struct{}{}
		log.Println("Request to stop the server")
	}

	<-stop
}
