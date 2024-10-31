package file

type Config struct {
	NutServers []NutServer `yaml:"nut_servers"`
}

type NutServer struct {
	Name     string         `yaml:"name"`
	Host     string         `yaml:"host"`
	Port     int            `yaml:"port"`
	Username string         `yaml:"username"`
	Password string         `yaml:"password"`
	Targets  []TargetServer `yaml:"targets"`
}

type TargetServer struct {
	Name      string   `yaml:"name"`
	MAC       string   `yaml:"mac"`
	Broadcast string   `yaml:"broadcast"`
	Port      int      `yaml:"port"`
	Rules     []string `yaml:"rules"`
}
