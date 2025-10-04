package ups

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	nut "github.com/robbiet480/go.nut"
)

type UPS struct {
	nut.Client
}

func connect(host string, port int, username, password string) (UPS, error) {
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
		return UPS{}, errors.New("authentication failed")
	}
	return UPS{client}, nil
}

func GetJSON(ns *entity.NutServer) (string, error) {
	client, err := connect(ns.Host, ns.Port, ns.Username, ns.Password)
	if err != nil {
		return "", fmt.Errorf("could not connect to NUT server: %w", err)
	}
	defer func(client *UPS) {
		_, err = client.Disconnect()
		if err != nil {
			log.Printf("Could not disconnect from NUT server: %s", err)
		}
	}(&client)

	inputJSON, err := client.toJSON()
	if err != nil {
		return "", fmt.Errorf("could not get UPS list: %w", err)
	}
	return inputJSON, nil
}

func (u *UPS) toJSON() (string, error) {
	ups, err := u.GetUPSList()
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(ups)
	if err != nil {
		return "", err
	}

	return string(jsonData), err
}
