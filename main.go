package main

import (
	"flag"
	"log"
)

const (
	ERROR_EXIT_CODE = 1
	working_table   = "jobs"
)

func init() {
	log.SetPrefix("<ActiveServer>: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Active Server")
	create := flag.Bool("create", false, "drop an existing JOBS table and create a new one")

	flag.Parse()

	as := NewActiveServer(uint16(*port))
	as.DB.NewConnection()
	if *create {
		if as.DB.TableExists(working_table) {
			if as.DB.DropTable(working_table) {
				as.DB.CreateTable(working_table)
			}
		}
	}

	log.Printf("Active server is running on http://0.0.0.0:%d", port)
	as.Run()
}
