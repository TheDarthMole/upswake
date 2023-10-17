package config

type Host struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Name        string        `yaml:"name"`
	Credentials []Credentials `yaml:"credentials"`
}

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type WakeHosts struct {
	Name      string  `yaml:"name"`
	Mac       string  `yaml:"mac"`
	Broadcast string  `yaml:"broadcast"`
	Port      int     `yaml:"port"`
	NutHost   NutHost `yaml:"nutHost"`
}

type NutHost struct {
	Name     string `yaml:"name"`
	username string `yaml:"username"`
}
