package entity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func genEcdsaKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func genRsaKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func generateTestCert(t *testing.T, priv, public any) ([]byte, error) {
	t.Helper()

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
		},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, public, priv)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func TestCLIArgs_Address(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "HTTP Port 8080 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
		{
			name: "HTTPS Port 8443 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "https://127.0.0.1:8443",
		},
		{
			name: "HTTPS Port 8443 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "https://1.2.3.4:8443",
		},
		{
			name: "HTTP Port 8080 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://1.2.3.4:8080",
		},
		{
			name: "HTTP Port 8080 on 0.0.0.0",
			fields: fields{
				Host:   net.ParseIP("0.0.0.0"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
		{
			name: "HTTP Port 8080 on ::",
			fields: fields{
				Host:   net.ParseIP("::"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if got := c.Address(); got != tt.want {
				t.Errorf("Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIArgs_URLPrefix(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test HTTP URL Prefix",
			fields: fields{
				UseSSL: false,
			},
			want: "http://",
		},
		{
			name: "Test HTTPS URL Prefix",
			fields: fields{
				UseSSL: true,
			},
			want: "https://",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if got := c.URLPrefix(); got != tt.want {
				t.Errorf("URLPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIArgs_Validate(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid HTTP Config",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8080",
				TLSConfig:  nil,
			},
			wantErr: false,
		},
		{
			name: "Valid HTTPS Config",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			wantErr: false,
		},
		{
			name: "HTTPS Config without Certfile",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			wantErr: true,
		},
		{
			name: "HTTPS Config without Keyfile",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			wantErr: true,
		},
		{
			name: "HTTPS Config without TLSConfig",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  nil,
			},
			wantErr: true,
		},
		{
			name: "HTTP Config without valid host",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("999.999.999.999"),
				Port:       "8080",
				TLSConfig:  nil,
			},
			wantErr: true,
		},
		{
			name: "HTTP Config with non-integer port",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "invalidPort",
				TLSConfig:  nil,
			},
			wantErr: true,
		},
		{
			name: "HTTP Config with port too large",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "999999999",
				TLSConfig:  nil,
			},
			wantErr: true,
		},
		{
			name: "HTTP Config with port too small",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "-1",
				TLSConfig:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCLIArgs_x509Cert(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	rsaKey, err := genRsaKey()
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	ecdsaKey, err := genEcdsaKey()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}
	rsaCertPEM, err := generateTestCert(t, rsaKey, &rsaKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}
	ecdsaCertPEM, err := generateTestCert(t, ecdsaKey, &ecdsaKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}
	if err = afero.WriteFile(fileSystem, "rsaServer.cert", rsaCertPEM, 0644); err != nil {
		t.Fatalf("failed to write to file system: %v", err)
	}
	if err = afero.WriteFile(fileSystem, "ecdsaServer.cert", ecdsaCertPEM, 0644); err != nil {
		t.Fatalf("failed to write to file system: %v", err)
	}
	if err = afero.WriteFile(fileSystem, "invalidServer.cert", []byte("invalid cert"), 0644); err != nil {
		t.Fatalf("failed to write to file system: %v", err)
	}
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid RSA Cert",
			fields: fields{
				CertFile: "rsaServer.cert",
			},
			wantErr: false,
		},
		{
			name: "Valid ecdsa Cert",
			fields: fields{
				CertFile: "ecdsaServer.cert",
			},
			wantErr: false,
		},
		{
			name: "Invalid Cert",
			fields: fields{
				CertFile: "invalidServer.cert",
			},
			wantErr: true,
		},
		{
			name: "Cert file does not exist",
			fields: fields{
				CertFile: "doesNotExist.cert",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			_, err := c.x509Cert(fileSystem)
			if (err != nil) != tt.wantErr {
				t.Errorf("x509Cert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewCLIArgs(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	type args struct {
		configFile string
		useSSL     bool
		certFile   string
		keyFile    string
		host       string
		port       string
	}
	tests := []struct {
		name    string
		args    args
		want    *CLIArgs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCLIArgs(fileSystem, tt.args.configFile, tt.args.useSSL, tt.args.certFile, tt.args.keyFile, tt.args.host, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCLIArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCLIArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
