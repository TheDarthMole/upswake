package repository

//go:generate mockgen -package mocks -source cli.go -destination mocks/cli_mock.go CLIArgsRepository

type CLIArgsRepository interface {
	Validate() error
	URLPrefix() string
	Address() string
	ListenAddress() string
	URL() string
}
