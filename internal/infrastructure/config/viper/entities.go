package viper

type Config struct {
	NutServers []*NutServer `mapstructure:"nut_servers"`
}

type NutServer struct {
	Name     string          `mapstructure:"name" json:"name"`
	Host     string          `mapstructure:"host" json:"host"`
	Username string          `mapstructure:"username" json:"username"`
	Password string          `mapstructure:"password" json:"password"`
	Targets  []*TargetServer `mapstructure:"targets" json:"targets"`
	Port     int             `mapstructure:"port" json:"port"`
}

type TargetServer struct {
	Name      string   `mapstructure:"name" json:"name"`
	MAC       string   `mapstructure:"mac" json:"mac"`
	Broadcast string   `mapstructure:"broadcast" json:"broadcast"`
	Interval  string   `mapstructure:"interval" json:"interval" default:"15m"`
	Rules     []string `mapstructure:"rules" json:"rules"`
	Port      int      `mapstructure:"port" json:"port" default:"9"`
}
