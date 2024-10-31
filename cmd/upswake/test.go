package main

import (
	config "github.com/TheDarthMole/UPSWake/internal/infrastructure/config/file"
	"log"
)

func main() {
	configs, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(configs)
}
