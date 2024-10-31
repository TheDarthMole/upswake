package entity

type Config struct {
	NutServers []NutServer
}

type NutServer struct {
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	Targets  []TargetServer
}

type TargetServer struct {
	Name      string
	MAC       string
	Broadcast string
	Port      int
	Rules     []string
}

type ConfigLoader interface {
	Load() (*Config, error)
}
