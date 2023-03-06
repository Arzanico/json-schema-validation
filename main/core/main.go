package main

import (
	"json-schema-validation/internal/server"
	"log"
)

func main() {
	srv := server.NewHttpServer(":8080")
	log.Println("Server is running ...")
	log.Fatal(srv.ListenAndServe())
}
