package repository

import (
	"crypto/tls"
)

type CLIArgsRepository interface {
	Validate() error
	x509Cert() (*tls.Config, error)
	URLPrefix() string
	Address() string
}
