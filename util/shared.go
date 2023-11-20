package util

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/ups"
)

func GetJSON(woLTarget *config.WoLTarget) (string, error) {
	ns := woLTarget.NutServer
	client, err := ups.Connect(ns.Host, ns.GetPort(), ns.Credentials.Username, ns.Credentials.Password)
	if err != nil {
		return "", fmt.Errorf("could not connect to NUT server: %s", err)
	}
	defer client.Disconnect()

	inputJson, err := client.ToJson()
	if err != nil {
		return "", fmt.Errorf("could not get UPS list: %s", err)
	}
	return inputJson, nil
}
