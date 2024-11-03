package main

import (
	config "github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"log"
)

func main() {
	configs, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(configs)
}
