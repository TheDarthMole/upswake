package repository

type CLIArgsRepository interface {
	Validate() error
	URLPrefix() string
	Address() string
	ListenAddress() string
	URL() string
}
