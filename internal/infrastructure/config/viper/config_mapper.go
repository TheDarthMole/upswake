package viper

import (
	"errors"
	"fmt"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
)

var ErrFailedParsingInterval = errors.New("failed to parse interval, must be a valid duration string")

func FromFileConfig(config *Config) (*entity.Config, error) {
	nutServers := make([]*entity.NutServer, len(config.NutServers))
	for i, nutServer := range config.NutServers {
		entityNutServer, err := FromFileNutServer(nutServer)
		if err != nil {
			return nil, err
		}
		nutServers[i] = entityNutServer
	}

	return &entity.Config{
		NutServers: nutServers,
	}, nil
}

func ToFileConfig(entityConfig *entity.Config) *Config {
	nutServers := make([]*NutServer, len(entityConfig.NutServers))
	for i, nutServer := range entityConfig.NutServers {
		nutServers[i] = ToFileNutServer(nutServer)
	}

	return &Config{
		NutServers: nutServers,
	}
}

func FromFileNutServer(nutServer *NutServer) (*entity.NutServer, error) {
	targets := make([]*entity.TargetServer, len(nutServer.Targets))
	for i, target := range nutServer.Targets {
		entityTarget, err := FromFileTargetServer(target)
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

func ToFileNutServer(nutServer *entity.NutServer) *NutServer {
	targets := make([]*TargetServer, len(nutServer.Targets))
	for i, target := range nutServer.Targets {
		targets[i] = ToFileTargetServer(target)
	}
	return &NutServer{
		Name:     nutServer.Name,
		Host:     nutServer.Host,
		Port:     nutServer.Port,
		Username: nutServer.Username,
		Password: nutServer.Password,
		Targets:  targets,
	}
}

func FromFileTargetServer(targetServer *TargetServer) (*entity.TargetServer, error) {
	interval, err := time.ParseDuration(targetServer.Interval)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedParsingInterval, err)
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

func ToFileTargetServer(targetServer *entity.TargetServer) *TargetServer {
	return &TargetServer{
		Name:      targetServer.Name,
		MAC:       targetServer.MAC,
		Broadcast: targetServer.Broadcast,
		Port:      targetServer.Port,
		Interval:  targetServer.Interval.String(),
		Rules:     targetServer.Rules,
	}
}
