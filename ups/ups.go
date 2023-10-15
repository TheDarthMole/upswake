package ups

import (
	"fmt"
	nut "github.com/robbiet480/go.nut"
	"log"
)

type UPS struct {
	nut.Client
}

func Connect(host, username, password string) (UPS, error) {
	client, err := nut.Connect(host)
	if err != nil {
		return UPS{}, err
	}

	authenticate, err := client.Authenticate(username, password)
	if err != nil {
		return UPS{}, err
	}
	if authenticate != true {
		log.Printf("Authentication failed to host '%s' as user '%s'", host, username)
		return UPS{}, err
	}
	return UPS{client}, nil
}

func (u *UPS) GetBatteryCharge(upsName string) (int64, error) {
	batteryCharge, err := u.getValueFromUPS(upsName, "battery.charge")
	if err != nil {
		return 0, err
	}
	return batteryCharge.(int64), nil
}

func (u *UPS) getValueFromUPS(upsName, variableName string) (interface{}, error) {
	list, err := u.GetUPSList()
	if err != nil {
		return "", err
	}
	for _, ups := range list {
		if ups.Name == upsName {
			for _, variable := range ups.Variables {
				if variable.Name == variableName {
					return variable.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("could not find UPS with name '%s'", upsName)
}
