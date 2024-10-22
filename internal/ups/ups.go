package ups

import (
	"encoding/json"
	"fmt"
	"github.com/TheDarthMole/UPSWake/internal/config"
	nut "github.com/robbiet480/go.nut"
	"log"
)

type UPS struct {
	nut.Client
}

func Connect(host string, port int, username, password string) (UPS, error) {
	client, err := nut.Connect(host, port)
	if err != nil {
		return UPS{}, err
	}

	authenticate, err := client.Authenticate(username, password)
	if err != nil {
		return UPS{}, err
	}
	if !authenticate {
		log.Printf("Authentication failed to host '%s' as user '%s'", host, username)
		return UPS{}, fmt.Errorf("authentication failed")
	}
	return UPS{client}, nil
}

func GetJSON(ns *config.NutServer) (string, error) {
	client, err := Connect(ns.Host, ns.GetPort(), ns.Credentials.Username, ns.Credentials.Password)
	if err != nil {
		return "", fmt.Errorf("could not connect to NUT server: %s", err)
	}
	defer func(client *UPS) {
		_, err := client.Disconnect()
		if err != nil {
			log.Printf("Could not disconnect from NUT server: %s", err)
		}
	}(&client)

	inputJson, err := client.ToJson()
	if err != nil {
		return "", fmt.Errorf("could not get UPS list: %s", err)
	}
	return inputJson, nil
}

func (u *UPS) ToJson() (string, error) {
	upss, err := u.GetUPSList()
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(upss)
	if err != nil {
		return "", err
	}

	return string(jsonData), err
}
