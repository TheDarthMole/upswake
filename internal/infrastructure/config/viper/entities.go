package viper

type Config struct {
	NutServers []NutServer `mapstructure:"nut_servers"`
}

type NutServer struct {
	Name     string          `mapstructure:"name"`
	Host     string          `mapstructure:"host"`
	Port     int             `mapstructure:"port"`
	Username string          `mapstructure:"username"`
	Password string          `mapstructure:"password"` //nolint:gosec // G117: TODO investigate secure ways to handle this password, such as using environment variables
	Targets  []*TargetServer `mapstructure:"targets"`
}

type TargetServer struct {
	Name      string   `mapstructure:"name"`
	MAC       string   `mapstructure:"mac"`
	Broadcast string   `mapstructure:"broadcast"`
	Port      int      `mapstructure:"port" default:"9"`
	Interval  string   `mapstructure:"interval" default:"15m"`
	Rules     []string `mapstructure:"rules"`
}
