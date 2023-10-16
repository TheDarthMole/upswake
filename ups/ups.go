package ups

import (
	"encoding/json"
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
	ups, err := u.getUPSFromList(upsName)
	if err != nil {
		return "", err
	}
	for _, variable := range ups.Variables {
		if variable.Name == variableName {
			return variable.Value, nil
		}
	}
	return "", fmt.Errorf("could not find UPS with name '%s'", upsName)
}

func (u *UPS) getUPSFromList(upsName string) (nut.UPS, error) {
	list, err := u.GetUPSList()
	if err != nil {
		return nut.UPS{}, fmt.Errorf("could not get UPS list: %w", err)
	}
	for _, ups := range list {
		if ups.Name == upsName {
			return ups, nil
		}
	}
	return nut.UPS{}, fmt.Errorf("could not find UPS with name '%s'", upsName)
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
