package main

import (
	"flag"
	"log"
)

const (
	ERROR_EXIT_CODE = 1
)

func init() {
	log.SetPrefix("<ActiveServer>: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Active Server")

	flag.Parse()

	app := NewActiveServer(uint16(*port))
	app.DB.NewConnection()
	log.Printf("Active server is running on http://0.0.0.0:%d", port)
	app.Run()
}
