package ups

import (
	nut "github.com/robbiet480/go.nut"
	"log"
)

func Connect(host, username, password string) (nut.Client, error) {
	client, err := nut.Connect(host)
	if err != nil {
		return nut.Client{}, err
	}

	authenticate, err := client.Authenticate(username, password)
	if err != nil {
		return nut.Client{}, err
	}
	if authenticate != true {
		log.Printf("Authentication failed to host '%s' as user '%s'", host, username)
		return nut.Client{}, err
	}
	return client, nil
}
