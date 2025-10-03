package entity

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/TheDarthMole/UPSWake/internal/util"
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
			return fmt.Errorf("SSL is enabled but certFile or keyFile is not set")
		}
		if c.TLSConfig == nil {
			return errors.New("TLSConfig cannot be null")
		}
	}

	if c.Host == nil {
		return errors.New("invalid listen host IP address")
	}

	portInt, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", err)
	}
	if portInt <= 0 || portInt > 65535 {
		return fmt.Errorf("invalid listen port %d", portInt)
	}
	return nil
}

func (c *CLIArgs) x509Cert(fileSystem afero.Fs) (*tls.Config, error) {
	certFile, err := util.GetFile(fileSystem, c.CertFile)
	if err != nil {
		return nil, err
	}

	// Decode the PEM certificate
	data, _ := pem.Decode(certFile)
	if data == nil {
		return nil, errors.New("failed to parse PEM certificate")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(data.Bytes)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{}
	conf.RootCAs = x509.NewCertPool()
	conf.RootCAs.AddCert(cert)

	return conf, nil
}

func (c *CLIArgs) URLPrefix() string {
	if c.UseSSL {
		return "https://"
	}
	return "http://"
}

func (c *CLIArgs) Address() string {
	baseURL := fmt.Sprintf("%s%s:%s", c.URLPrefix(), c.Host.String(), c.Port)
	if c.Host.IsUnspecified() {
		baseURL = fmt.Sprintf("%s127.0.0.1:%s", c.URLPrefix(), c.Port)
	}

	return baseURL
}
