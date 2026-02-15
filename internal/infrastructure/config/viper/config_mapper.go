package viper

import "github.com/TheDarthMole/UPSWake/internal/domain/entity"

func fromFileConfig(config *Config) *entity.Config {
	nutServers := make([]*entity.NutServer, len(config.NutServers))
	for i, nutServer := range config.NutServers {
		nutServers[i] = fromFileNutServer(&nutServer)
	}

	return &entity.Config{
		NutServers: nutServers,
	}
}

func fromFileNutServer(nutServer *NutServer) *entity.NutServer {
	targets := make([]*entity.TargetServer, len(nutServer.Targets))
	for i, target := range nutServer.Targets {
		targets[i] = fromFileTargetServer(target)
	}

	return &entity.NutServer{
		Name:     nutServer.Name,
		Host:     nutServer.Host,
		Port:     nutServer.Port,
		Username: nutServer.Username,
		Password: nutServer.Password,
		Targets:  targets,
	}
}

func fromFileTargetServer(targetServer *TargetServer) *entity.TargetServer {
	return &entity.TargetServer{
		Name:      targetServer.Name,
		MAC:       targetServer.MAC,
		Broadcast: targetServer.Broadcast,
		Port:      targetServer.Port,
		Interval:  targetServer.Interval,
		Rules:     targetServer.Rules,
	}
}
