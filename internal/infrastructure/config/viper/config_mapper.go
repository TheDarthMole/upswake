package viper

import (
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
)

func fromFileConfig(config *Config) (*entity.Config, error) {
	nutServers := make([]*entity.NutServer, len(config.NutServers))
	for i, nutServer := range config.NutServers {
		entityNutServer, err := fromFileNutServer(&nutServer)
		if err != nil {
			return nil, err
		}
		nutServers[i] = entityNutServer
	}

	return &entity.Config{
		NutServers: nutServers,
	}, nil
}

func fromFileNutServer(nutServer *NutServer) (*entity.NutServer, error) {
	targets := make([]*entity.TargetServer, len(nutServer.Targets))
	for i, target := range nutServer.Targets {
		entityTarget, err := fromFileTargetServer(target)
		if err != nil {
			return nil, err
		}
		targets[i] = entityTarget
	}

	return &entity.NutServer{
		Name:     nutServer.Name,
		Host:     nutServer.Host,
		Port:     nutServer.Port,
		Username: nutServer.Username,
		Password: nutServer.Password,
		Targets:  targets,
	}, nil
}

func fromFileTargetServer(targetServer *TargetServer) (*entity.TargetServer, error) {
	interval, err := time.ParseDuration(targetServer.Interval)
	if err != nil {
		return nil, err
	}

	return &entity.TargetServer{
		Name:      targetServer.Name,
		MAC:       targetServer.MAC,
		Broadcast: targetServer.Broadcast,
		Port:      targetServer.Port,
		Interval:  interval,
		Rules:     targetServer.Rules,
	}, nil
}
