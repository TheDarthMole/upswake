package entity

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/afero"
)

type CLIArgs struct {
	ConfigFile string
	UseSSL     bool
	CertFile   string
	KeyFile    string
	Host       net.IP
	Port       string
	TLSConfig  *tls.Config
}

var (
	ErrCertFilesNotSet    = errors.New("SSL is enabled but certFile or keyFile is not set")
	ErrTLSConfigNotSet    = errors.New("TLSConfig cannot be null")
	ErrFailedParsePEM     = errors.New("failed to parse PEM certificate")
	ErrFailedReadCertFile = errors.New("failed to read certificate file")
)

func NewCLIArgs(fileSystem afero.Fs, configFile string, useSSL bool, certFile, keyFile, host, port string) (*CLIArgs, error) {
	cliArgs := &CLIArgs{
		ConfigFile: configFile,
		UseSSL:     useSSL,
		CertFile:   certFile,
		KeyFile:    keyFile,
		Host:       net.ParseIP(host),
		Port:       port,
	}
	if useSSL {
		tlsConfig, err := cliArgs.x509Cert(fileSystem)
		if err != nil {
			return nil, err
		}
		cliArgs.TLSConfig = tlsConfig
	}
	err := cliArgs.Validate()
	if err != nil {
		return nil, err
	}
	return cliArgs, nil
}

func (c *CLIArgs) Validate() error {
	if c.UseSSL {
		if c.CertFile == "" || c.KeyFile == "" {
			return ErrCertFilesNotSet
		}
		if c.TLSConfig == nil {
			return ErrTLSConfigNotSet
		}
	}

	if c.Host == nil {
		return ErrHostRequired
	}

	portInt, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPort, err)
	}
	if portInt <= 0 || portInt > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidPort, portInt)
	}
	return nil
}

func (c *CLIArgs) x509Cert(fileSystem afero.Fs) (*tls.Config, error) {
	certFile, err := afero.ReadFile(fileSystem, c.CertFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedReadCertFile, err)
	}

	// Decode the PEM certificate
	data, _ := pem.Decode(certFile)
	if data == nil {
		return nil, ErrFailedParsePEM
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(data.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedParsePEM, err)
	}

	conf := &tls.Config{
		RootCAs: x509.NewCertPool(),
	}
	conf.RootCAs.AddCert(cert)

	return conf, nil
}

func (c *CLIArgs) URLPrefix() string {
	if c.UseSSL {
		return "https://"
	}
	return "http://"
}

func (c *CLIArgs) address(unspecifiedHost string) string {
	host := c.Host.String()
	if c.Host.IsUnspecified() {
		host = unspecifiedHost
	}
	return net.JoinHostPort(host, c.Port)
}

func (c *CLIArgs) Address() string {
	return c.address("127.0.0.1")
}

func (c *CLIArgs) ListenAddress() string {
	return c.address("::")
}

func (c *CLIArgs) URL() string {
	return fmt.Sprintf("%s%s", c.URLPrefix(), c.Address())
}
