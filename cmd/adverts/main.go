package main

import (
	"log"

	"github.com/ineverbee/adverts-go-api/internal/app"
)

func main() {
	err := app.StartServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server stopped!")
}
