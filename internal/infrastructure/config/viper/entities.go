package viper

type Config struct {
	NutServers []NutServer `mapstructure:"nut_servers"`
}

type NutServer struct {
	Name     string          `mapstructure:"name"`
	Host     string          `mapstructure:"host"`
	Username string          `mapstructure:"username"`
	Password string          `mapstructure:"password"`
	Targets  []*TargetServer `mapstructure:"targets"`
	Port     int             `mapstructure:"port"`
}

type TargetServer struct {
	Name      string   `mapstructure:"name"`
	MAC       string   `mapstructure:"mac"`
	Broadcast string   `mapstructure:"broadcast"`
	Interval  string   `mapstructure:"interval" default:"15m"`
	Rules     []string `mapstructure:"rules"`
	Port      int      `mapstructure:"port" default:"9"`
}
